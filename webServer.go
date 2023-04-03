package main

import (
	"net/http"
)

type WebHttp struct {
}

func (h *WebHttp) Index(writer http.ResponseWriter, request *http.Request) error {
	//os.ReadFile()
	return nil
}
