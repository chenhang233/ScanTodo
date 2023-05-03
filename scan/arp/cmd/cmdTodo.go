package main

import (
	"ScanTodo/scan/arp"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

// op := flag.Uint("op", 1, "操作 1请求2回复")
// sIp := flag.String("sIp", "", "发送方ip")
// sMAC := flag.String("sMAC", "", "发送方MAC")
// tIp := flag.String("tIp", "", "接收方ip")
// tMAC := flag.String("tMAC", "", "接收方MAC")
// timeout := flag.Duration("t", time.Hour, "最大超时时间")
// interval := flag.Duration("i", time.Minute, "间隔时间")
//
//	flag.Usage = func() {
//		print(usage)
//		flag.PrintDefaults()
//	}
//
// flag.Parse()
//
// if flag.NArg() == 0 {
// fmt.Println("最后一个参数本机指定网卡IP")
// flag.Usage()
// return
// }
var usage = `
Usage:
	-load arp.json
`

const ConfigPath = "configs"

type Config struct {
	configInfo []ConfigInfo
}

type ConfigInfo struct {
	Op     uint16        `json:"op"`
	T      time.Duration `json:"t"`
	I      time.Duration `json:"i"`
	SIP    string        `json:"sIP"`
	SMAC   string        `json:"sMAC"`
	TIP    string        `json:"tIP"`
	TMAC   string        `json:"tMAC"`
	HostIp string        `json:"hostIp"`
}

var config ConfigInfo

func init() {
	load := flag.String("load", "arp.json", "读取配置文件")
	flag.Usage = func() {
		print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	var err error
	p := os.O_RDWR | os.O_CREATE | os.O_APPEND
	_, err = os.Stat(ConfigPath)
	if err != nil {
		err = os.MkdirAll(ConfigPath, os.ModePerm)
	}
	goOut(err)
	_, err = os.OpenFile(ConfigPath+"/"+*load, p, os.ModePerm)
	goOut(err)
	file, err := os.ReadFile(ConfigPath + "/" + *load)
	goOut(err)
	err = json.Unmarshal(file, &config)
	goOut(err)
	if config.I == time.Duration(0) {
		config.I = time.Minute
	}
	if config.T == time.Duration(0) {
		config.T = time.Hour
	}
}

func goOut(err error) {
	if err != nil {
		flag.Usage()
		panic(err)
	}
}

func main() {
	host := config.HostIp
	m, err := arp.New(host)
	if err != nil {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			m.Log.Debug.Println("接收到中断的信号:", i)
			m.Stop()
		}
	}()

	m.Operation = config.Op
	m.Timeout = config.T
	m.Interval = config.I
	m.SourceDevice.Ip = net.ParseIP(config.SIP)
	hw1, err := net.ParseMAC(config.SMAC)
	if err != nil {
		m.Log.Warn.Println(err)
		return
	}
	m.SourceDevice.Mac = hw1
	m.TargetDevice.Ip = net.ParseIP(config.TIP)
	hw2, err := net.ParseMAC(config.TMAC)
	if err != nil {
		m.Log.Warn.Println(err)
		return
	}
	m.TargetDevice.Mac = hw2
	s1 := fmt.Sprintf("\n本机信息(ip: %v,mac: %v,Description: %s)", m.SelfDevice.Ip, m.SelfDevice.Mac, m.SelfDevice.Description)
	s2 := fmt.Sprintf("\n源信息(ip: %v,mac: %v,Description: %s)", m.SourceDevice.Ip, m.SourceDevice.Mac, m.SourceDevice.Description)
	s3 := fmt.Sprintf("\n目标信息(ip: %v,mac: %v,Description: %s)", m.TargetDevice.Ip, m.TargetDevice.Mac, m.TargetDevice.Description)
	m.Log.Info.Println(s1 + s2 + s3)
	m.OnSend = func(m *arp.Metadata) {
		fmt.Println("发送ARP报文")
	}
	m.OnFinish = func(m *arp.Metadata) {
		fmt.Println("结束")
	}
	err = m.Run()
	if err != nil {
		m.Log.Debug.Println("异常结束日志:", err)
	}
}
