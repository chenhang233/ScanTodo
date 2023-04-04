package main

import (
	"log"
	"net/http"
	"os"
)

type WebHttp struct {
}

func (h *WebHttp) Index(writer http.ResponseWriter, request *http.Request) {
	file, err := os.ReadFile("./index.html")
	if err != nil {
		log.Println("读文件错误", err)
	}
	writer.Header().Set("Content-Type", "text/html")
	switch request.Method {
	case "GET":
		_, err := writer.Write(file)
		if err != nil {
			log.Println(err)
		}
	}
}
