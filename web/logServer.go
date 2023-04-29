package web

import (
	"ScanTodo/scanLog"
	"encoding/json"
	"net/http"
)

type LogHttp struct {
	Log *scanLog.LogConf
}

func (h *LogHttp) Index(writer http.ResponseWriter, request *http.Request) {
	jr := &JsonResponse{Code: NormalCode, Message: ParamsErrorMsg, Data: nil}
	js, _ := json.Marshal(jr)
	writer.Write(js)
}
func (h *LogHttp) Tcp(writer http.ResponseWriter, request *http.Request) {
	//body := request.Body
	//all, err := utils.MyReadAll(body)
	//if err != nil {
	//	h.Log.Error.Println("body Read", err)
	//}
	//utils.GetLogName()
}
func (h *LogHttp) Icmp(writer http.ResponseWriter, request *http.Request) {

}
func (h *LogHttp) Ws(w http.ResponseWriter, r *http.Request) {

}
