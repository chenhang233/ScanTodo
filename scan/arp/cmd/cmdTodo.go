package main

import (
	"ScanTodo/scan/arp"
	"encoding/json"
	"flag"
	"fmt"
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

var config arp.ConfigInfo

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
		config.I = time.Second * 5
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
	m, err := arp.New(&config)
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			m.Log.Debug.Println("An interrupt signal was received:", i)
			m.Stop()
		}
	}()

	s1 := fmt.Sprintf("\nSelfDevice(ip: %v,mac: %v,Description: %s)", m.SelfDevice.Ip, m.SelfDevice.Mac, m.SelfDevice.Description)
	s2 := fmt.Sprintf("\nSourceDevice(ip: %v,mac: %v,Description: %s)", m.SourceDevice.Ip, m.SourceDevice.Mac, m.SourceDevice.Description)
	s3 := fmt.Sprintf("\nTargetDevice(ip: %v,mac: %v,Description: %s)", m.TargetDevice.Ip, m.TargetDevice.Mac, m.TargetDevice.Description)
	m.Log.Info.Println(s1 + s2 + s3)

	f := m.EnableRuleIpTcp
	if f {
		rules := m.IpTcpRules
		s4 := fmt.Sprintf("\nSourceIps: %v,SourcePorts: %v", rules.SourceIps, rules.SourcePorts)
		s5 := fmt.Sprintf("\nDestinationIps: %v,DestinationPorts: %v", rules.DestinationIps, rules.DestinationPorts)
		m.Log.Info.Println(s4 + s5)
	}
	m.OnSetup = func() {

	}
	m.OnSend = func(m *arp.Metadata) {
		fmt.Println(fmt.Sprintf("arp OnSend: %v ---> %v", m.SourceDevice.Mac, m.TargetDevice.Mac))
	}
	m.OnFinish = func(m *arp.Metadata) {
		fmt.Println("OnFinish")
	}
	err = m.Run()
	if err != nil {
		m.Log.Debug.Println("Abnormal end log:", err)
	}
}
