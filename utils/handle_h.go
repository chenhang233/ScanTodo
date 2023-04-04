package utils

import (
	"log"
	"os"
)

type Methods string

var MethodsMap = map[string]string{
	"GET":  "GET",
	"POST": "POST",
}

func HandleHttpMethod(key string) {
	method, f := MethodsMap[key]
	if !f {
		log.Panicln("http请求方法错误")
		os.Exit(1)
	}
	switch method {
	case "GET":
	case "POST":
	}
}
