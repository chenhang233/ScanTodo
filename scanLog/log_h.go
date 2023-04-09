package scanLog

import (
	"io"
	"log"
	"os"
	"time"
)

const (
	logsFilePath = "scanLog/logs"
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
	_, err2 := os.Stat(logsFilePath)
	if err2 != nil {
		os.MkdirAll(logsFilePath, os.ModePerm)
	}
	format := time.Now().Format("_2006_01_02")
	logFile, err := os.OpenFile(logsFilePath+"/"+conf.ServiceName+format+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
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
