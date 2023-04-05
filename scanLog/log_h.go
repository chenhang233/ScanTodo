package scanLog

import (
	"io"
	"log"
	"os"
	"time"
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
	format := time.Now().Format("_2006_01_02")
	logFile, err := os.OpenFile("scanLog/"+conf.ServiceName+format+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
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
