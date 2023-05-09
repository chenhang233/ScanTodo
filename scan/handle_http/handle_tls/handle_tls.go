package handle_tls

type Metadata struct {
	ContentType string
	Version     string
	Length      byte
}

var TlsVersionMap map[byte]string

func (m *Metadata) init() {
	if TlsVersionMap != nil {
		return
	}
	TlsVersionMap = map[byte]string{}
	TlsVersionMap[00] = "SSL 3.0"
	TlsVersionMap[01] = "TLS 1.0"
	TlsVersionMap[02] = "TLS 1.1"
	TlsVersionMap[03] = "TLS 1.2"
	TlsVersionMap[04] = "TLS 1.3"

}

func IsTlsProtocol(bys []byte) (bool, Metadata) {
	tls := Metadata{}
	tls.init()
	if len(bys) < 5 {
		return false, tls
	}
	t := bys[0]
	if t == 22 {
		tls.ContentType = "handshake"
	} else if t == 23 {
		tls.ContentType = "application Data"
	} else {
		return false, tls
	}
	vb := bys[1:3]
	if vb[0] != 03 {
		return false, tls
	}
	version, ok := TlsVersionMap[vb[1]]
	if !ok {
		return false, tls
	}
	tls.Version = version
	tls.Length = bys[4:5][0]
	return true, tls
}
