package main

import (
	"ScanTodo/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	NoMessageCode = 10
	NoMessageMsg  = "你干嘛 哎呦~"
)

type JsonResponse struct {
	code    int
	message string
	data    interface{}
}

type JsonRequest struct {
	data interface{}
}

type TcpReq struct {
	ip      string
	port    string
	timeout string
}

type WebHttp struct {
}

func (h *WebHttp) Index(writer http.ResponseWriter, request *http.Request) {
	file, err := os.ReadFile("./index.html")
	if err != nil {
		log.Println("读文件错误", err)
	}
	writer.Header().Set("Content-Type", "text/html")
	utils.HandleHttpMethod("D")
	switch request.Method {
	case "GET":
		writer.Write(file)
	case "POST":
		jr := &JsonResponse{code: NoMessageCode, message: NoMessageMsg, data: nil}
		js, err := json.Marshal(jr)
		if err != nil {
			log.Println("序列化错误", err)
		}
		writer.Write(js)
	}
}

func (h *WebHttp) Tcp(writer http.ResponseWriter, request *http.Request) {
	body := request.Body
	req := &JsonRequest{data: &TcpReq{}}
	bys := make([]byte, 0)
	body.Read(bys)
	fmt.Println(string(bys), "bys")
	json.Unmarshal(bys, req)
}
