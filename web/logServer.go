package web

import "net/http"

type LogHttp struct {
}

func (*LogHttp) Index(http.ResponseWriter, *http.Request) {

}
func (*LogHttp) Tcp(http.ResponseWriter, *http.Request) {

}
func (*LogHttp) Icmp(http.ResponseWriter, *http.Request) {

}
func (*LogHttp) Ws(w http.ResponseWriter, r *http.Request) {

}
