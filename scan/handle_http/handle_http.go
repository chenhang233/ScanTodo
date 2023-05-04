package handle_http

import (
	"ScanTodo/scanLog"
	"errors"
	"fmt"
	"strings"
)

type Metadata struct {
	Row  string
	Head map[string]string
	Body string
}

func (m *Metadata) ReadHTTP(bys []byte, log *scanLog.LogConf) error {
	var err error
	var readHTTPNewLine func() (string, error)
	readHTTPNewLine = func() (string, error) {
		if bys[0] == '\r' && bys[1] == '\n' {
			bys = bys[2:]
			return "", nil
		}
		sb := strings.Builder{}
		for i, b := range bys {
			if b == '\r' && b+1 == '\n' {
				bys = bys[i+2:]
				return sb.String(), nil
			}
			sb.WriteByte(b)
		}
		return sb.String(), errors.New(fmt.Sprintf("readHTTPNewLine 解析错误: %s, 数据: %v", sb, bys))
	}
	m.Row, err = readHTTPNewLine()
	if err != nil {
		log.Warn.Println(err)
		return err
	}
	var line string
	for {
		line, err = readHTTPNewLine()
		if err != nil {
			log.Warn.Println(err)
		}
		if line == "" {
			break
		}
		sp := strings.Split(line, ":")
		if len(sp) > 2 {
			e := fmt.Sprintf("请求行或响应行切割错误: %v", line)
			log.Warn.Println(e)
			err = errors.New(e)
		}
		m.Head[sp[0]] = sp[1]
	}
	m.Body = string(bys)
	return err
}
