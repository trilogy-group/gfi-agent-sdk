/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package logger

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/trilogy-group/gfi-agent-sdk/logger/lumberjack"
	"github.com/trilogy-group/gfi-agent-sdk/version"
)

var Logger *logger

const GFIAgentLogDir = "C:\\ProgramData\\GFIAgent\\Logs"

func LogDir() string {
	return GFIAgentLogDir
}

type LogFile struct {
	Name      string `json:"file_name"`
	Size      int64  `json:"size"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type logger struct {
	mu         sync.Mutex
	fileWriter *logrus.Logger
	stdWriter  *logrus.Logger
}

func GetFileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func GetStdFields() *logrus.Fields {
	proc := GetFileNameWithoutExtension(filepath.Base(os.Args[0]))
	return &logrus.Fields{
		"version": version.Long(),
		"proc":    proc,
		"pid":     os.Getpid(),
	}
}

func (L *logger) Info(v ...interface{}) {

	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Info(v...)

	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Info(v...)

}

func (L *logger) Warning(v ...interface{}) {
	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Warning(v...)
	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Warning(v...)
}

func (L *logger) Error(v ...interface{}) {
	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Error(v...)
	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Warning(v...)

}

func (L *logger) Infof(format string, v ...interface{}) {
	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Infof(format, v...)

	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Infof(format, v...)
}

func (L *logger) Warningf(format string, v ...interface{}) {
	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Warningf(format, v...)
	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Warningf(format, v...)
}

func (L *logger) Errorf(format string, v ...interface{}) {
	L.mu.Lock()
	L.fileWriter.WithFields(*GetStdFields()).Errorf(format, v...)
	updateIndexFile()
	L.mu.Unlock()

	L.stdWriter.WithFields(*GetStdFields()).Errorf(format, v...)
}

func GetStartTime(file string) (time.Time, error) {
	inFile, err := os.Open(filepath.Join(LogDir(), file))

	if err != nil {
		return time.Now(), err
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		log := scanner.Text()
		return time.Parse("2006-01-02 15:04:05", log[0:19])
	}
	// empty file
	return time.Now(), nil
}

func List(dir string) ([]fs.FileInfo, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	return f.Readdir(0)
}

func ListFiles(dir string, ext string) ([]string, error) {
	files, err := List(dir)
	if err != nil {
		return nil, err
	}
	dirs := []string{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ext {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func ListLogFiles() ([]LogFile, error) {
	if logFiles, err := ListFiles(LogDir(), ".log"); err != nil {
		return nil, err
	} else {
		res := []LogFile{}
		for _, file := range logFiles {
			if stat, err := os.Stat(filepath.Join(LogDir(), file)); err != nil {
				return nil, err
			} else {
				logFile := LogFile{
					Name:    file,
					Size:    stat.Size(),
					EndTime: stat.ModTime().Format("2006-01-02T15:04:05Z"),
				}
				if startTime, err := GetStartTime(file); err != nil {
					return nil, err
				} else {
					logFile.StartTime = startTime.Format("2006-01-02T15:04:05Z")
					res = append(res, logFile)
				}
			}
		}
		return res, nil
	}
}

func init() {
	Logger = &logger{mu: sync.Mutex{}}
	offsetData = 0
	fullLogFilename = filepath.Join(LogDir(), "gfiagent.log")
	fullIndexFilename = filepath.Join(LogDir(), "gfiagent.log.idx")

	lumberjackLogrotate := &lumberjack.Logger{
		Filename:   filepath.Join(LogDir(), "gfiagent.log"),
		MaxSize:    50, // Max megabytes before log is rotated
		MaxBackups: 10, // Max number of old log files to keep
		MaxAge:     7,  // Max number of days to retain log files
		Compress:   true,
		OnRotate: func() {
			err := os.Remove(fullIndexFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			createLogIndexFile()
		},
	}

	Logger.fileWriter = &logrus.Logger{
		Out:   lumberjackLogrotate,
		Level: logrus.InfoLevel,
		Formatter: &nested.Formatter{
			ShowFullLevel:   true,
			NoColors:        true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}

	Logger.stdWriter = &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Formatter: &nested.Formatter{
			ShowFullLevel:   true,
			NoColors:        true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}

	hasIdxFile, err := exists(fullIndexFilename)
	if !hasIdxFile && err == nil {
		createLogIndexFile()
	} else {
		readLastOffset()
	}
}
