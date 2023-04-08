package main

import (
	"ScanTodo/utils"
	"flag"
	"fmt"
	"net/http"
)

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")

type WebService interface {
	Index(http.ResponseWriter, *http.Request)
	Tcp(http.ResponseWriter, *http.Request)
}

func main() {
	var sever WebService
	sever = &WebHttp{}
	flag.Parse()
	utils.HubInstance = utils.NewHub()
	go utils.HubInstance.Run()
	http.HandleFunc("/", sever.Index)
	http.HandleFunc("/tcp", sever.Tcp)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		utils.ServeWs(utils.HubInstance, w, r)
	})
	fmt.Println("服务器启动完成...")
	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(err)
	}
}
