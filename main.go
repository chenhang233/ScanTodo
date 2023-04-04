package main

import (
	"ScanTodo/utils"
	"fmt"
	"net/http"
)

type WebService interface {
	Index(http.ResponseWriter, *http.Request)
	Tcp(http.ResponseWriter, *http.Request)
}

func main() {
	var sever WebService
	sever = &WebHttp{}
	http.HandleFunc("/", sever.Index)
	http.HandleFunc("/tcp", sever.Tcp)
	fmt.Println("服务器启动完成...")
	utils.HandleHttpMethod("GET")
	utils.HandleHttpMethod("GA")
	if err := http.ListenAndServe("127.0.0.1:8000", nil); err != nil {
		panic(err)
	}
}
