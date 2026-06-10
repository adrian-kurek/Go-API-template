package log

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

var (
	pathToLogDirectory = "../../tmp/"
	dateFormat         = "2006.01.02"
	timeFormat         = "15:04:05"
)

func removeDirectoryAfterTest() {
	err := os.RemoveAll(pathToLogDirectory)
	if err != nil {
		fmt.Println(err)
	}
}

func removeFile(t *testing.T, actualDate string) {
	err := os.Remove(pathToLogDirectory + actualDate + ".json")
	if err != nil {
		t.Fatal(err)
	}
}

type log struct {
	Date      string
	TypeOfLog string
	Message   string
	Data      any
}

func TestLoggerInfo(t *testing.T) {
	type args struct {
		name               string
		message            string
		expectedContentLen int
		expectedTypeOfLog  string
	}

	testsCases := []args{
		{
			name:               "Log with message",
			message:            "Test string",
			expectedContentLen: 211,
			expectedTypeOfLog:  "INFO",
		},
	}

	for _, testCase := range testsCases {
		t.Run(testCase.name, func(t *testing.T) {
			logger := NewLogger(pathToLogDirectory, dateFormat, timeFormat)
			logger.Info(testCase.message, "")
			logger.Close()

			actualDate := time.Now().Format(dateFormat)
			content, _ := os.ReadFile(pathToLogDirectory + actualDate + ".json")

			var logs []log
			err := json.Unmarshal(content, &logs)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, testCase.expectedContentLen, len(content))
			for i := 1; i < len(logs); i++ {
				assert.Equal(t, testCase.expectedTypeOfLog, logs[i].TypeOfLog)
				assert.Equal(t, testCase.message, logs[i].Message)
			}
		})
	}
	removeDirectoryAfterTest()
}

func TestLoggerWarning(t *testing.T) {
	type args struct {
		name               string
		message            string
		expectedContentLen int
		expectedTypeOfLog  string
	}

	testsCases := []args{
		{
			name:               "Log with message",
			message:            "Test string",
			expectedContentLen: 214,
			expectedTypeOfLog:  "WARNING",
		},
	}

	for _, testCase := range testsCases {
		t.Run(testCase.name, func(t *testing.T) {
			logger := NewLogger(pathToLogDirectory, dateFormat, timeFormat)
			logger.Warning(testCase.message, "")
			logger.Close()

			actualDate := time.Now().Format(dateFormat)
			content, _ := os.ReadFile(pathToLogDirectory + actualDate + ".json")

			var logs []log
			err := json.Unmarshal(content, &logs)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, testCase.expectedContentLen, len(content))
			for i := 1; i < len(logs); i++ {
				assert.Equal(t, testCase.expectedTypeOfLog, logs[i].TypeOfLog)
				assert.Equal(t, testCase.message, logs[i].Message)
			}
		})
	}
	removeDirectoryAfterTest()
}

func TestLoggerError(t *testing.T) {
	type args struct {
		name               string
		message            string
		expectedContentLen int
		expectedTypeOfLog  string
	}

	testsCases := []args{
		{
			name:               "Log with message",
			message:            "Test string",
			expectedContentLen: 212,
			expectedTypeOfLog:  "ERROR",
		},
	}

	for _, testCase := range testsCases {
		t.Run(testCase.name, func(t *testing.T) {
			logger := NewLogger(pathToLogDirectory, dateFormat, timeFormat)
			logger.Error(testCase.message, "")
			logger.Close()

			actualDate := time.Now().Format(dateFormat)
			content, _ := os.ReadFile(pathToLogDirectory + actualDate + ".json")

			var logs []log
			err := json.Unmarshal(content, &logs)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, testCase.expectedContentLen, len(content))
			for i := 1; i < len(logs); i++ {
				assert.Equal(t, testCase.expectedTypeOfLog, logs[i].TypeOfLog)
				assert.Equal(t, testCase.message, logs[i].Message)
			}
		})
	}
	removeDirectoryAfterTest()
}

func TestLoggerValidate(t *testing.T) {
	type args struct {
		name               string
		message            string
		expectedContentLen int
		expectedTypeOfLog  string
	}
	testsCases := []args{
		{
			name:               "Properly created new file due to day change",
			message:            "Test",
			expectedContentLen: 116,
			expectedTypeOfLog:  "INFO",
		},
	}

	for _, testCase := range testsCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDate := time.Now().Format(dateFormat)

			logger := NewLogger(pathToLogDirectory, dateFormat, timeFormat)

			removeFile(t, actualDate)

			logger.startTime = time.Now().Add(time.Hour * time.Duration(-24)).Format(dateFormat)
			logger.validate()
			logger.Close()

			content, _ := os.ReadFile(pathToLogDirectory + actualDate + ".json")

			var logs []log
			err := json.Unmarshal(content, &logs)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, testCase.expectedContentLen, len(content))
			for i := 1; i < len(logs); i++ {
				assert.Equal(t, testCase.expectedTypeOfLog, logs[i].TypeOfLog)
				assert.Equal(t, testCase.message, logs[i].Message)
			}
		})
	}
	removeDirectoryAfterTest()
}