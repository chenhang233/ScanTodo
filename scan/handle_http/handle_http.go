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
			if b == '\r' && bys[i+1] == '\n' {
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
	keyM := map[string]string{}
	for {
		line, err = readHTTPNewLine()
		if err != nil {
			log.Warn.Println(err)
		}
		if line == "" {
			m.Head = keyM
			break
		}
		sp, err := m.Split(line, ':')
		if err != nil {
			log.Warn.Println(err)
		}
		keyM[sp[0]] = sp[1]
	}
	m.Body = string(bys)
	return err
}

func (m *Metadata) Split(line string, separator byte) (sp [2]string, e error) {
	for i, v := range line {
		if v > 255 {
			e = errors.New(fmt.Sprintf("请求行或响应行切割错误: %v", line))
			return
		}
		b := byte(v)
		if b == separator {
			sp = [2]string{}
			sp[0] = line[:i]
			sp[1] = line[i+1:]
			break
		}
	}
	return
}
