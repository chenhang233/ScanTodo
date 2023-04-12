package main

import (
	"ScanTodo/scan"
	"ScanTodo/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const (
	NormalCode      = 0
	NoMessageCode   = 10 // 没有这个功能
	NoMessageMsg    = "你干嘛 哎呦~"
	ParamsErrorCode = 11 // 参数错误
	ParamsErrorMsg  = "参数错误"
)

type JsonResponse struct {
	Code    int
	Message string
	Data    interface{}
}

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
		writer.Write(file)
	case "POST":
		jr := &JsonResponse{Code: NoMessageCode, Message: NoMessageMsg, Data: nil}
		js, _ := json.Marshal(jr)
		writer.Write(js)
	}
}

func (h *WebHttp) Tcp(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		body := request.Body
		all, err := utils.MyReadAll(body)
		if err != nil {
			log.Panicln("body Read", err)
		}

		sc, _ := scan.NewScanCase("TCP", all)
		ctx := context.WithValue(context.Background(), "tcp", "1")
		err = sc.Repo.Start(ctx)
		jr := JsonResponse{Code: NormalCode, Message: "无"}
		if err != nil {
			sc.Log.Error.Println("tcp start error", err)
			jr.Message = err.Error()
			jr.Code = ParamsErrorCode
		}
		js, _ := json.Marshal(jr)
		writer.Write(js)
		sc.Repo.End(ctx)
	}
}

func (h *WebHttp) Icmp(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		body := request.Body
		all, err := utils.MyReadAll(body)
		if err != nil {
			log.Panicln("body Read", err)
		}

		sc, _ := scan.NewScanCase("ICMP", all)
		ctx := context.WithValue(context.Background(), "icmp", "1")
		err = sc.Repo.Start(ctx)
		jr := JsonResponse{Code: NormalCode, Message: "等待消息"}
		if err != nil {
			sc.Log.Error.Println("icmp start error", err)
			jr.Message = err.Error()
			jr.Code = ParamsErrorCode
		}
		js, _ := json.Marshal(jr)
		writer.Write(js)
		sc.Repo.End(ctx)
	}
}

func (h *WebHttp) Ws(writer http.ResponseWriter, request *http.Request) {
	utils.ServeWs(utils.HubInstance, writer, request)
}
