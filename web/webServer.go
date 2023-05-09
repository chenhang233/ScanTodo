package web

import (
	"ScanTodo/scan"
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

type WebHttp struct {
	Log *scanLog.LogConf
}

func (h *WebHttp) Index(writer http.ResponseWriter, request *http.Request) {
	openUrl := "web/index.html"
	if request.RequestURI == "/favicon.ico" {
		return
	}
	if request.RequestURI != "/index" {
		openUrl = "web" + request.RequestURI
	}
	file, err := os.ReadFile(openUrl)
	if err != nil {
		h.Log.Error.Println("读文件错误", err)
	}
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
			h.Log.Error.Println("body Read", err)
		}

		sc, _ := scan.NewScanCase("TCP", all)
		ctx := context.WithValue(context.Background(), "tcp", "1")
		err = sc.Repo.Start(ctx)
		jr := JsonResponse{Code: NormalCode, Message: "结束"}
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
			h.Log.Error.Println("body Read", err)
		}

		sc, _ := scan.NewScanCase("ICMP", all)
		ctx := context.WithValue(context.Background(), "icmp", "1")
		err = sc.Repo.Start(ctx)
		jr := JsonResponse{Code: NormalCode, Message: "结束"}
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

func (h *WebHttp) Arp(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		body := request.Body
		all, err := utils.MyReadAll(body)
		if err != nil {
			h.Log.Error.Println("body Read", err)
		}

		sc, _ := scan.NewScanCase("ARP", all)
		ctx := context.WithValue(context.Background(), "icmp", "1")
		err = sc.Repo.Start(ctx)
		jr := JsonResponse{Code: NormalCode, Message: "结束"}
		if err != nil {
			sc.Log.Error.Println("arp start error", err)
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
