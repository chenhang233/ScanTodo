package scanLog

import (
	"io"
	"log"
	"os"
	"time"
)

const (
	logsFilePath = "scanLog/logs"
	HTTPLogPath  = "HTTPLogPath"
	ScansLogPath = "ScanLogPath"
	PingLogPath  = "PingLogPath"
)

type LogConf struct {
	ServiceName string
	Debug       *log.Logger
	Info        *log.Logger
	Warn        *log.Logger
	Error       *log.Logger
}

func LoadLog(sName string) (*LogConf, error) {
	conf := &LogConf{
		ServiceName: sName,
	}
	err := handleDir()
	if err != nil {
		return nil, err
	}
	format := time.Now().Format("2006_01_02")
	logFile, err := os.OpenFile(logsFilePath+"/"+conf.ServiceName+"/"+format+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	conf.Debug = log.New(multiWriter, "[debug]", log.Ldate|log.Ltime|log.Lshortfile)
	conf.Info = log.New(multiWriter, "[info]", log.Ldate|log.Ltime|log.Lshortfile)
	conf.Warn = log.New(multiWriter, "[warn]", log.Ldate|log.Ltime|log.Lshortfile)
	conf.Error = log.New(multiWriter, "[error]", log.Ldate|log.Ltime|log.Lshortfile)
	return conf, nil
}

func handleDir() error {
	var err error
	_, err = os.Stat(logsFilePath)
	if err != nil {
		err = os.MkdirAll(logsFilePath, os.ModePerm)
	}
	logPathMap := make(map[int]string, 3)
	logPathMap[0] = logsFilePath + "/" + HTTPLogPath
	logPathMap[1] = logsFilePath + "/" + ScansLogPath
	logPathMap[2] = logsFilePath + "/" + PingLogPath
	for _, m := range logPathMap {
		_, err = os.Stat(m)
		if err != nil {
			err = os.MkdirAll(m, os.ModePerm)
		}
	}
	return err
}
