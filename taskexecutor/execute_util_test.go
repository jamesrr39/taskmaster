package taskexecutor

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getFormattedTime(t *testing.T) {
	assert.Equal(t, "03:04:05.006", getFormattedTime(mockNowProvider()))
}

func Test_writeStringToLogFile(t *testing.T) {
	byteBuffer := bytes.NewBuffer(nil)

	writeStringToLogFile("task finished successfully", byteBuffer, "STDOUT", mockNowProvider)

	assert.Equal(t, "03:04:05.006: STDOUT: task finished successfully\n", string(byteBuffer.Bytes()))
}

func Test_writeToLogFile(t *testing.T) {
	reader := bytes.NewBuffer(nil)
	writer := bytes.NewBuffer(nil)
	sourceName := "STDOUT"

	_, err := reader.WriteString("task finished successfully")
	assert.Nil(t, err)

	err = writeToLogFile(reader, writer, sourceName, mockNowProvider)
	assert.Nil(t, err)

	assert.Equal(t, "03:04:05.006: STDOUT: task finished successfully\n", string(writer.Bytes()))
}

func mockNowProvider() time.Time {
	nSec := 6 * 1000 * 1000
	date := time.Date(2000, 1, 2, 3, 4, 05, nSec, time.UTC)
	return date
}
