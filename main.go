package main

import "net/http"

type WebService interface {
	Index(http.ResponseWriter, *http.Request) error
}

func main() {

}


