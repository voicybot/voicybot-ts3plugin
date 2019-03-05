package main

import (
	"encoding/binary"
	"io"
	"regexp"
)

// readSamples reads a raw 16-bit signed 2-channel PCM audio stream and converts
// the given amount of samples to 16-bit integer values. The output does not
// necessarily match the given wanted sample count if reading ends early!
func readSamples(r io.Reader, sampleCount int) (output []int16, nSamples int, err error) {
	bytesToRead := 2 * sampleCount // read 16 bit values
	bytesArray := make([]byte, bytesToRead)
	n, err := r.Read(bytesArray)
	if err != nil {
		return
	}
	readValueCount := n / 2
	// nSamples = readValueCount / 2
	// log(fmt.Sprintf("nSamples: %d", nSamples), teamlog.Debug)
	output = make([]int16, readValueCount)
	var unsignedValue uint16
	for i := 0; i < readValueCount; i++ {
		//binary.Read(r, binary.LittleEndian, &sample)
		unsignedValue = binary.LittleEndian.Uint16(bytesArray[i*2 : (i+1)*2])
		output[i] = int16(unsignedValue)
	}
	return
}

// bbTagsRegex matches all BBTag markers, start and end.
var bbTagsRegex = regexp.MustCompile(`\[/?[A-Za-z0-9_]+\]`)

// stripBBTags removes all BBTag markers, start and end, from a text.
func stripBBTags(text string) string {
	return bbTagsRegex.ReplaceAllString(text, "")
}

// downmixSingleSampleToMono takes 16-bit values for a single sample and mixes it
// down to a single Mono sample by calculating the average of all values.
func downmixSingleSampleToMono(values ...int16) int16 {
	var sum int = 0
	for _, v := range values {
		sum += int(v)
	}
	return int16(sum / len(values))
}

// downmixMultipleSamplesToMono takes a sequence of samples and the channel count
// to mix all samples down to a mono sample sequence.
func downmixMultipleSamplesToMono(samples []int16, channels int) (newSamples []int16) {
	newSamples = make([]int16, len(samples)/channels)
	for i := 0; i < len(samples); i++ {
		oldSample := samples[i : i+channels]
		newSamples[i/channels] = downmixSingleSampleToMono(oldSample...)
	}
	return
}

// amplifySamples takes all given samples and amplifies them with the given
// ratio. A ratio of 1 means no change, anything lower will decrease and anything
// higher will increase the volume.
func amplifySamples(samples []int16, ratio float64) (newSamples []int16) {
	newSamples = make([]int16, len(samples))
	for i, oldSample := range samples {
		newSamples[i] = int16(float64(oldSample) * ratio)
	}
	return
}
