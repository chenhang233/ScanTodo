package main

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"ScanTodo/web"
	"flag"
	"fmt"
	"net/http"
)

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")

type WebService interface {
	Index(http.ResponseWriter, *http.Request)
	Tcp(http.ResponseWriter, *http.Request)
	Icmp(http.ResponseWriter, *http.Request)
	Arp(http.ResponseWriter, *http.Request)
	Ws(w http.ResponseWriter, r *http.Request)
}

type MainService struct {
	Log *scanLog.LogConf
	w   WebService
}

func main() {
	loadLog, err := scanLog.LoadLog(scanLog.HTTPLogPath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	loadLog.Debug.Println("全局日志开始启动.............................")
	ms := &MainService{
		Log: loadLog,
		w: &web.WebHttp{
			Log: loadLog,
		},
	}
	ms.Log.Debug.Println("服务启动中,全局日志加载完成")
	flag.Parse()

	utils.HubInstance = utils.NewHub()
	ms.Log.Debug.Println("服务启动中,websocket实例初始化完成")
	go utils.HubInstance.Run()
	ms.Log.Debug.Println("服务启动中,websocket开启监听完成")
	http.HandleFunc("/", ms.w.Index)
	http.HandleFunc("/tcp", ms.w.Tcp)
	http.HandleFunc("/icmp", ms.w.Icmp)
	http.HandleFunc("/arp/proxy", ms.w.Arp)
	http.HandleFunc("/ws", ms.w.Ws)
	ms.Log.Debug.Println("服务启动成功: ", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		ms.Log.Error.Println("服务启动失败,", err)
		panic(err)
	}
}
