package web

const (
	NormalCode      = 0
	NoMessageCode   = 10 // 没有这个功能
	NoMessageMsg    = "你干嘛 哎呦~"
	ParamsErrorCode = 11 // 参数错误
	ParamsErrorMsg  = "参数错误"
)

type JsonResponse struct {
	Code    int
	Message string
	Data    interface{}
}
