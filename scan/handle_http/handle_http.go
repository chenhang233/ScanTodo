package handle_http

import "strings"

type Metadata struct {
	Row  string
	Head string
	Body string
}

func (m *Metadata) ReadHTTP(bys []byte) error {
	var readHTTPNewLine func() string
	readHTTPNewLine = func() string {
		sb := strings.Builder{}
		for i, b := range bys {
			if b == '\r' && b+1 == '\n' {
				bys = bys[i+2:]
				return sb.String()
			}
			sb.WriteByte(b)
		}
		return ""
	}
	m.Row = readHTTPNewLine()
	head := strings.Builder{}
	for {
		line := readHTTPNewLine()
		if line == "" {
			break
		} else {
			head.WriteString(line)
		}
	}
	return nil
}
