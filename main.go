package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	ts3plugin "github.com/icedream/go-ts3plugin"
	"github.com/icedream/go-ts3plugin/teamlog"
	"github.com/icedream/go-ts3plugin/teamspeak"
	"github.com/voicybot/voicybot-ts3plugin/resolvers"

	// Resolvers
	_ "github.com/voicybot/voicybot-ts3plugin/resolvers/youtubedl"
)

const (
	Name    = "Voicybot"
	Author  = "Carl Kittelberger"
	Version = "0.1.0"
)

var volume float64 = 0.5

func catchPanic() {
	if err := recover(); err != nil {
		if ts3plugin.Functions() != nil {
			log(fmt.Sprintf("%s\n%s",
				err,
				string(debug.Stack())), teamlog.Critical)
		}
	}
}

func log(msg string, severity teamlog.LogLevel) {
	ts3plugin.Functions().LogMessage(
		msg,
		severity, Name, 0)
}

func logDebug(msg string, args ...interface{}) {
	log(fmt.Sprintf(msg, args...), teamlog.Debug)
}

func logInfo(msg string, args ...interface{}) {
	log(fmt.Sprintf(msg, args...), teamlog.Info)
}

func logWarn(msg string, args ...interface{}) {
	log(fmt.Sprintf(msg, args...), teamlog.Warning)
}

func logError(msg string, args ...interface{}) {
	log(fmt.Sprintf(msg, args...), teamlog.Error)
}

func logCritical(msg string, args ...interface{}) {
	log(fmt.Sprintf(msg, args...), teamlog.Critical)
}

func init() {
	ts3plugin.Name = Name
	ts3plugin.Author = Author
	ts3plugin.Version = Version

	// ts3plugin.OnUserLoggingMessageEvent = func(logMessage string, logLevel teamlog.LogLevel, logChannel string, logID uint64, logTime time.Time, completeLogString string) {
	// 	// Print all log messages to the current chat tab
	// 	ts3plugin.Functions().PrintMessageToCurrentTab(
	// 		fmt.Sprintf("[COLOR=gray][I]%s[/I]\t%s\t[B]%s[/B]\t%s[/COLOR]",
	// 			logTime,
	// 			logLevel,
	// 			logChannel,
	// 			logMessage))
	// }

	ts3plugin.Init = func() (ok bool) {
		defer catchPanic()

		// youtubedl_resolver.RegisterResolver()

		logDebug("Plugin init")

		ok = true
		return
	}

	ts3plugin.Shutdown = func() {
		defer catchPanic()

		// youtubedl_resolver.UnregisterResolver()

		logDebug("Plugin shutdown")

		stopPlayback(0) // FIXME !!!!!!
	}

	ts3plugin.OnTextMessageEvent = func(
		serverConnectionHandlerID uint64,
		targetMode teamspeak.AnyID,
		toID teamspeak.AnyID,
		fromID teamspeak.AnyID,
		fromName string,
		fromUniqueIdentifier string,
		message string,
		ffIgnored bool) (retCode int) {
		fields := strings.Fields(stripBBTags(message))
		defer catchPanic()

		switch fields[0] {
		case "volume":
			if len(fields) < 2 {
				ts3plugin.Functions().RequestSendPrivateTextMsg(
					serverConnectionHandlerID,
					"You need at least a volume value (0 to 100).",
					fromID,
					"")
				return
			}
			f, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				ts3plugin.Functions().RequestSendPrivateTextMsg(
					serverConnectionHandlerID,
					err.Error(),
					fromID,
					"")
				return
			}
			if f > 100 || f < 0 {
				ts3plugin.Functions().RequestSendPrivateTextMsg(
					serverConnectionHandlerID,
					"Volume value needs to be in valid range (0 to 100).",
					fromID,
					"")
				return
			}
			volume = f / 100
			logDebug("Volume is now: %v", volume)

		case "play":
			go func() {
				defer catchPanic()
				if len(fields) < 2 {
					ts3plugin.Functions().PrintMessageToCurrentTab("You need at least the URL to play back.")
					return
				}
				uri := fields[1]
				var videoPassword string
				if len(fields) > 2 {
					videoPassword = fields[2]
				}
				if err := setUpPlayback(serverConnectionHandlerID, uri, videoPassword); err != nil {
					ts3plugin.Functions().RequestSendPrivateTextMsg(
						serverConnectionHandlerID,
						fmt.Sprintf("Can not start playback of %s: %s", fields[1], err.Error()),
						fromID,
						"")
					return
				}
			}()

		case "stop":
			go func() {
				defer catchPanic()
				err := stopPlayback(serverConnectionHandlerID)
				if err != nil {
					ts3plugin.Functions().RequestSendPrivateTextMsg(
						serverConnectionHandlerID,
						fmt.Sprintf("Can not stop playback: %s", err.Error()),
						fromID,
						"")
					return
				}
				ts3plugin.Functions().RequestSendPrivateTextMsg(
					serverConnectionHandlerID,
					"Stopped playback.",
					fromID,
					"")
			}()
		}

		return 0
	}

	ts3plugin.OnEditCapturedVoiceDataEvent = func(serverConnectionHandlerID uint64, samples *ts3plugin.Samples, isMuted bool) (shouldMute bool) {
		defer catchPanic()

		shouldMute = isMuted

		if currentOutputStream == nil || !isPlaybackRunning {
			//log("No audio stream, skipping", teamlog.Debug)
			return
		}

	loop:
		for i := 0; i < samples.Channels()*samples.SampleCount(); i += samples.Channels() {
			select {
			case newSample, ok := <-currentOutputStream:
				if !ok {
					// channel has been closed
					stopPlayback(serverConnectionHandlerID)
					break loop
				}
				newSample = amplifySamples(newSample, volume)
				switch samples.Channels() {
				case 2:
					samples.SetSamples(i, newSample)
				case 1:
					samples.SetSamples(
						i,
						[]int16{downmixSingleSampleToMono(newSample...)})
				default:
					log("Unsupported channel count.", teamlog.Warning)
					stopPlayback(serverConnectionHandlerID)
					return
				}
			default:
				//log(fmt.Sprintf("Not enough samples, leaving rest at %d.", i), teamlog.Debug)
				break loop
			}
		}
		shouldMute = shouldMute && !samples.IsEdited()
		return
	}
}

var currentOutputStream <-chan []int16
var currentInputStream io.ReadCloser
var playbackMutex sync.Mutex
var isPlaybackRunning = false

func setUpPlayback(serverConnectionHandlerID uint64, uri string, videoPassword string) (err error) {
	playbackMutex.Lock()
	defer playbackMutex.Unlock()

	// check for valid URL
	u, err := url.Parse(uri)
	if err != nil {
		logDebug("Parsing url failed: %s", err.Error())
		return
	}

	if isPlaybackRunning {
		err = errors.New("There is already a running playback that needs to be stopped first.")
		return
	}

	// attempt to resolve
	var errorStream io.ReadCloser
	inputStream, errorStream, err := resolvers.Resolve(u, videoPassword)
	if err != nil {
		return
	}
	if inputStream == nil {
		err = errors.New("This URL could not be resolved. Please check if you spelled it correctly.")
		return
	}
	currentInputStream = inputStream
	go func() {
		defer catchPanic()
		logDebug("RESOLVER STDERR <listener loop started>")
		defer logDebug("RESOLVER STDERR <listener loop stopped>")
		r := bufio.NewReader(errorStream)
		for {
			item, err := r.ReadString('\n')
			if err != nil {
				logDebug("RESOLVER: <err: %s>", err.Error())
				break
			}
			for _, line := range strings.Split(item, "\r") {
				line = strings.Trim(line, "\r\n")
				logDebug("RESOLVER STDERR: %s", line)
			}
		}
	}()

	// start decoder
	logDebug("About to start new decoder...")
	decoder, err := newDecoder()
	if err != nil {
		return
	}
	decoder.LogStderr = func(msg string) {
		logDebug("FFMPEG STDERR: %s", msg)
	}
	go func() {
		defer catchPanic()
		logDebug("FFMPEG ERRCHAN <listener loop started>")
		defer logDebug("FFMPEG ERRCHAN <listener loop stopped>")
		for v := range decoder.ErrorC() {
			logError("FFMPEG ERRCHAN: %s", v.Error())
			stopPlayback(serverConnectionHandlerID)
			break
		}
	}()
	currentOutputStream = decoder.SamplesC()
	isPlaybackRunning = true

	go func() {
		defer catchPanic()
		defer inputStream.Close()
		defer decoder.Close()
		io.Copy(decoder, inputStream)
	}()

	// Needs to be disabled because otherwise the sound will not be let through by TeamSpeak
	ts3plugin.Functions().SetPreProcessorConfigValue(serverConnectionHandlerID, "vad", "false")

	return
}

func stopPlayback(serverConnectionHandlerID uint64) (err error) {
	playbackMutex.Lock()
	defer playbackMutex.Unlock()

	logDebug("Checking if playback is running")
	if !isPlaybackRunning {
		logDebug("Playback was already inactive")
		return
	}

	if currentInputStream != nil {
		logDebug("Closing current input stream")
		err = currentInputStream.Close()
		logDebug("Setting current input stream to nil")
		currentInputStream = nil
	}

	logDebug("Flushing output buffer...")
	for _ = range currentOutputStream {
	}

	logDebug("Notifying that playback is not running anymore")
	isPlaybackRunning = false

	ts3plugin.Functions().SetPreProcessorConfigValue(serverConnectionHandlerID, "vad", "true")

	return
}

// This will never be run!
func main() {
	fmt.Println("=======================================")
	fmt.Println("This is a TeamSpeak3 plugin, do not run this as a CLI application!")
	fmt.Println("Args were: ", os.Args)
	fmt.Println("=======================================")
}
