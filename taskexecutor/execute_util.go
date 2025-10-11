package taskexecutor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func getFormattedTime(now time.Time) string {
	milliSeconds := int(float64(now.Nanosecond()) / float64(1000000))
	return fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour(), now.Minute(), now.Second(), milliSeconds)
}

func writeStringToLogFile(text string, writer io.Writer, sourceName Source, nowProvider NowProvider) error {

	now := nowProvider()
	entry := LogEntry{
		Timestamp: Timestamp(now),
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

func writeToLogFile(pipe io.Reader, writer io.Writer, sourceName Source, nowProvider NowProvider) error {
	pipeScanner := bufio.NewScanner(pipe)
	for pipeScanner.Scan() {
		err := writeStringToLogFile(pipeScanner.Text(), writer, sourceName, nowProvider)
		// entry := LogEntry{
		// 	Timestamp: Timestamp(now),
		// 	Text:      pipeScanner.Text(),
		// 	Source:    sourceName,
		// }
		// err := json.NewEncoder(writer).Encode(entry)
		if nil != err {
			return err
		}
	}
	return nil
}

type Source string

const (
	SourceTaskmasterHarness Source = "TASKMASTER_HARNESS"
	SourceTaskmasterStdout  Source = "STDOUT"
	SourceTaskmasterStderr  Source = "STDERR"
)

type LogEntry struct {
	Timestamp Timestamp `json:"timestamp"`
	Text      string    `json:"text"`
	Source    Source    `json:"source"`
}
