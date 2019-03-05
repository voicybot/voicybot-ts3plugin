package main

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"
)

var (
	ffmpegBinaryName = "ffmpeg"
	ffmpegOptions    = []string{
		"-loglevel", "warning", // only print warnings and errors
		"-c:a", "pcm_s16le", // 16-bit signed PCM audio codec
		"-ar", "48000", // 48 kHz sample rate
		"-ac", "2", // stereo (we downmix to mono ourselves where necessary)
		"-f", "s16le", // use raw PCM audio for output
	}
)

// Decoder launches ffmpeg to transcode whatever media input into 16-bit signed
// PCM 48kHz audio, then feeds it out via a channel ready to inject into the
// TeamSpeak voice buffer.
type Decoder struct {
	ffmpeg    *exec.Cmd
	ffmpegIn  io.WriteCloser
	ffmpegOut io.ReadCloser
	ffmpegErr *bufio.Reader

	mutex  sync.Mutex
	closed bool

	errorC   <-chan error
	samplesC <-chan []int16

	LogStderr func(string)
}

func newDecoder() (decoder *Decoder, err error) {
	logDebug("Creating new decoder instance...")

	ffmpegArgs := []string{
		"-i", "-", // stdin for input
	}
	ffmpegArgs = append(ffmpegArgs, ffmpegOptions...)
	ffmpegArgs = append(ffmpegArgs,
		"pipe:", // output everything to stdout
	)
	ffmpeg := exec.Command(ffmpegBinaryName, ffmpegArgs...)
	ffmpegIn, err := ffmpeg.StdinPipe()
	if err != nil {
		return
	}
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		return
	}
	ffmpegErr, err := ffmpeg.StderrPipe()
	if err != nil {
		return
	}
	ffmpegErrBuffered := bufio.NewReader(ffmpegErr)

	err = ffmpeg.Start()
	if err != nil {
		logDebug("Starting ffmpeg failed: %s", err.Error())
		return
	}

	samplesC := make(chan []int16, 0.250*48000)
	errorC := make(chan error)

	decoder = &Decoder{
		ffmpeg:    ffmpeg,
		ffmpegIn:  ffmpegIn,
		ffmpegOut: ffmpegOut,
		ffmpegErr: ffmpegErrBuffered,
		samplesC:  samplesC,
		errorC:    errorC,
	}

	// stderr reading
	go func() {
		logDebug("Decoder stderr reading started")
		defer logDebug("Decoder stderr reading stopped")
		defer close(errorC)

		lineReader := bufio.NewReader(ffmpegErr)
		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				return
			}

			// remove any leftover line breaks
			line = strings.Trim(line, "\r\n")

			if decoder.LogStderr != nil {
				decoder.LogStderr(line)
			}
		}
	}()

	// stdout sample decoding loop
	go func() {
		logDebug("Decoder stdout sample decoding loop started")
		defer logDebug("Decoder stdout sample decoing loop stopped")
		defer decoder.Close()
		defer close(samplesC)

		// Just in case short reads do occur and we get less than the wanted
		// amount of samples, buffer them here.
		previousSampleArray := []int16{}

		for {
			sample, _, err := readSamples(ffmpegOut, 2-len(previousSampleArray))
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case errorC <- err:
				default:
				}
				return
			}
			sample = append(previousSampleArray, sample...)
			if len(sample) < 2 {
				continue
			}
			if len(sample) > 2 {
				select {
				case errorC <- errors.New("Got more than 2 short values for sample, logic error!"):
				default:
				}
				return
			}
			samplesC <- sample
		}
	}()

	return
}

func (decoder *Decoder) Write(p []byte) (n int, err error) {
	decoder.mutex.Lock()
	defer decoder.mutex.Unlock()
	return decoder.ffmpegIn.Write(p)
}

func (decoder *Decoder) Close() (err error) {
	decoder.mutex.Lock()
	defer decoder.mutex.Unlock()

	if decoder.closed {
		err = errors.New("Decoder was already closed")
		return
	}

	err = decoder.ffmpegIn.Close()
	decoder.ffmpeg.Wait()
	decoder.closed = true
	return
}

func (decoder *Decoder) SamplesC() <-chan []int16 {
	return decoder.samplesC
}

func (decoder *Decoder) ErrorC() <-chan error {
	return decoder.errorC
}
