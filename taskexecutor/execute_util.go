package taskexecutor

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

func getFormattedTime(now time.Time) string {
	milliSeconds := int(float64(now.Nanosecond()) / float64(1000000))
	return fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour(), now.Minute(), now.Second(), milliSeconds)
}

func writeStringToLogFile(text string, writer io.Writer, sourceName string, nowProvider NowProvider) error {
	_, err := writer.Write([]byte(getFormattedTime(nowProvider()) + ": " + sourceName + ": " + text + "\n"))
	if nil != err {
		return err
	}
	return nil
}

func writeToLogFile(pipe io.Reader, writer io.Writer, sourceName string, nowProvider NowProvider) error {
	pipeScanner := bufio.NewScanner(pipe)
	for pipeScanner.Scan() {
		_, err := writer.Write([]byte(getFormattedTime(nowProvider()) + ": " + sourceName + ": " + pipeScanner.Text() + "\n"))
		if nil != err {
			return err
		}
	}
	return nil
}
