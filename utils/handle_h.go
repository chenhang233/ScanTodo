package utils

import (
	"context"
	"io"
)

type Methods string

var MethodsMap = map[string]string{
	"GET":  "GET",
	"POST": "POST",
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

func MyReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func CheckIp() {

}

func CheckPost() {

}

type MyLog struct {
}

func (l *MyLog) LogError(ctx context.Context, msg string, e error) {

}
