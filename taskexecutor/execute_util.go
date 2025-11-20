package taskexecutor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jamesrr39/taskmaster/taskrunner"
)

func getFormattedTime(now time.Time) string {
	milliSeconds := int(float64(now.Nanosecond()) / float64(1000000))
	return fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour(), now.Minute(), now.Second(), milliSeconds)
}

func writeStringToLogFile(text string, writer io.Writer, sourceName SourceID, nowProvider NowProvider) error {

	now := nowProvider()
	entry := LogEntry{
		Timestamp: taskrunner.Timestamp(now),
		Text:      text,
		Source:    sourceName,
	}
	err := json.NewEncoder(writer).Encode(entry)
	// _, err := writer.Write([]byte(getFormattedTime(nowProvider()) + ": " + sourceName + ": " + text + "\n"))
	if nil != err {
		return err
	}
	return nil
}

func writeToLogFile(pipe io.Reader, writer io.Writer, sourceName SourceID, nowProvider NowProvider) error {
	pipeScanner := bufio.NewScanner(pipe)
	for pipeScanner.Scan() {
		err := writeStringToLogFile(pipeScanner.Text(), writer, sourceName, nowProvider)
		if nil != err {
			return err
		}
	}
	return nil
}

type SourceID int

const (
	SourceTaskmasterHarness SourceID = 1
	SourceTaskmasterStdout  SourceID = 2
	SourceTaskmasterStderr  SourceID = 3
)

var sourceNames = []string{
	"UNKNOWN",
	"TASKMASTER_HARNESS",
	"STDOUT",
	"STDERR",
}

func (s SourceID) String() string {
	return sourceNames[int(s)]
}

type LogEntry struct {
	Timestamp taskrunner.Timestamp `json:"timestamp"`
	Text      string               `json:"text"`
	Source    SourceID             `json:"source"`
}
