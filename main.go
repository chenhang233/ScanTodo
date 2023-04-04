package main

import (
	"fmt"
	"net/http"
)

type WebService interface {
	Index(http.ResponseWriter, *http.Request)
}

func main() {
	var sever WebService
	sever = &WebHttp{}
	http.HandleFunc("/", sever.Index)
	fmt.Println("服务器启动完成...")
	if err := http.ListenAndServe("127.0.0.1:8000", nil); err != nil {
		panic(err)
	}
}
