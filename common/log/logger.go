package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	reset  = "\x1b[0m"
)

type Logger struct {
	logDir     string
	dateFormat string
	timeFormat string
	file       *os.File
	startTime  string
}

func NewLogger(logDir string, dateFormat string, timeFormat string) *Logger {
	logger := &Logger{
		logDir:     logDir,
		dateFormat: dateFormat,
		timeFormat: timeFormat,
	}
	logger.initializeLogger()
	return logger
}

func (l *Logger) getLogTime() string {
	actualDate := time.Now()
	logTime := actualDate.Format(l.dateFormat + " " + l.timeFormat)
	return logTime
}

func (l *Logger) printDataToTheConsole(data ...any) {
	if len(data) > 0 {
		fmt.Print(" ")
		for _, d := range data {
			fmt.Print(d, " ")
		}
	}
}

func (l *Logger) printHeaderToTheConsole(message, typeOfLog, logTime string) {
	var color string
	switch typeOfLog {
	case "INFO":
		color = green
	case "WARNING":
		color = yellow
	case "ERROR":
		color = red
	}
	fmt.Println(color + "[" + typeOfLog + ": " + logTime + "] " + message)
}

func (l *Logger) printLogToTheConsole(message, typeOfLog, logTime string, data any) {
	l.printHeaderToTheConsole(message, typeOfLog, logTime)

	l.printDataToTheConsole(data)

	fmt.Println(reset)
}

func (l *Logger) emit(message, typeOfLog string, data any) {
	convertedDataToJSON, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	logTime := l.getLogTime()

	l.printLogToTheConsole(message, typeOfLog, logTime, data)

	fileContentToAdd := fmt.Sprintf(",{\n\t\"date\": \"%s\",\n\t\"typeOfLog\": \"%s\",\n\t\"message\": \"%s\",\n\t\"data\": %s\n}",
		logTime,
		typeOfLog,
		message,
		convertedDataToJSON,
	)

	_, err = l.file.WriteString(fileContentToAdd)
	if err != nil {
		fmt.Println("something went wrong during writing to data to the file")
	}
}

func (l *Logger) initializeLogger() {
	if _, err := os.Stat(l.logDir); os.IsNotExist(err) {
		if err := os.Mkdir(l.logDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	actualDate := time.Now()
	filename := actualDate.Format(l.dateFormat)
	l.startTime = filename

	file, err := os.OpenFile(l.logDir+"/"+filename+".json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		panic(err)
	}

	l.file = file
	stats, err := l.file.Stat()
	if err != nil {
		panic(err)
	}

	if stats.Size() > 0 {
		content, err := os.ReadFile(l.logDir + "/" + filename + ".json")
		if err != nil {
			panic(err)
		}
		lastBrace := bytes.LastIndexByte(content, '}')
		if lastBrace == -1 {
			l.file.WriteString("[")
		} else {
			os.Truncate(l.logDir+"/"+filename+".json", int64(lastBrace+1))
			_, err = l.file.WriteString(",")
			if err != nil {
				fmt.Println("something went wrong during writing to data to the file")
			}
		}
	} else {
		_, err = l.file.WriteString("[")
		if err != nil {
			fmt.Println("something went wrong during writing to data to the file")
		}
	}

	fileContentToAdd := fmt.Sprintf("{\n\t\"date\": \"%s\",\n\t\"typeOfLog\": \"%s\",\n\t\"message\": \"%s\",\n\t\"data\": %s\n}",
		l.getLogTime(), "INFO", "successfully created logger", []byte("{}"))

	_, err = l.file.WriteString(fileContentToAdd)
	if err != nil {
		fmt.Println("something went wrong during writing to data to the file")
	}
	l.Info("successfully initialized new logger", nil)
}

func (l *Logger) validate() {
	actualDate := time.Now().Format(l.dateFormat)

	if actualDate != l.startTime {

		fmt.Println("closing old file and creating the new one for new date")

		_, err := l.file.WriteString("]")
		if err != nil {
			fmt.Println("something went wrong during writing  data to the file")
		}

		err = l.file.Close()
		if err != nil {
			fmt.Println("something went wrong during writing  data to the file")
		}

		l.startTime = actualDate
		fileName := actualDate

		file, err := os.OpenFile(l.logDir+"/"+fileName+".json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			panic(err)
		}

		l.file = file
		_, err = l.file.WriteString("[")
		if err != nil {
			fmt.Println("something went wrong during writing  data to the file")
		}

		fileContentToAdd := fmt.Sprintf("{\n\t\"date\": \"%s\",\n\t\"typeOfLog\": \"%s\",\n\t\"message\": \"%s\",\n\t\"data\": %s\n}",
			l.getLogTime(), "INFO", "successfully created new file", []byte("{}"))

		_, err = l.file.WriteString(fileContentToAdd)
		if err != nil {
			fmt.Println("something went wrong during writing  data to the file")
		}
	}
}

func (l *Logger) Info(message string, data any) {
	l.validate()
	l.emit(message, "INFO", data)
}

func (l *Logger) Error(message string, data any) {
	l.validate()
	l.emit(message, "ERROR", data)
}

func (l *Logger) Warning(message string, data any) {
	l.validate()
	l.emit(message, "WARNING", data)
}

func (l *Logger) Close() error {
	if l.file == nil {
		fmt.Println("failed to close the file")
		return nil
	}
	_, err := l.file.WriteString("]")
	if err != nil {
		fmt.Println("something went wrong during writing  data to the file")
		return err
	}
	err = l.file.Close()
	if err != nil {
		return err
	}
	return nil
}
