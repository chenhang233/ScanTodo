package ping

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"errors"
	"github.com/google/uuid"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	timeSliceLength = 8
	trackerLength   = len(uuid.UUID{})
	protocolICMP    = 1
)

var (
	ipv4Proto = map[string]string{"icmp": "ip4:icmp", "udp": "udp4"}
	ipv6Proto = map[string]string{"icmp": "ip6:ipv6-icmp", "udp": "udp6"}
)

type Packet struct {
}

type Statistics struct{}

// Metadata 数据包元数据结构体
type Metadata struct {
	// 报文间隔时间
	Interval time.Duration
	// 报文超时时间
	Timeout time.Duration
	// 发送次数
	Count int
	// 已经发送包数量
	PacketsSent int
	// 已经收到包数量
	PacketsReceive int
	//收到重复包数量
	PacketsReceiveDuplicates int
	// 往返时间统计
	minRoundTripTime     time.Duration
	maxRoundTripTime     time.Duration
	averageRoundTripTime time.Duration
	// 读写锁
	statsMutex sync.RWMutex
	// 正在发送的包大小
	Size int
	// 源ip
	Source string
	// 所有往返时间统计
	RTTs []time.Duration
	// 监听完成时
	OnSetup func()
	//发送一个包后
	OnSend func(*Packet)
	//接收到一个包后
	OnReceive func(*Packet)
	// 接收到重复的包后
	OnDuplicateReceive func(*Packet)
	// 当ping结束后
	OnFinish func(*Statistics)
	// 发送数据包uuid列表
	trackerUUIDs []uuid.UUID
	id           int
	sequence     int
	// 记录序列号
	awaitingSequences map[uuid.UUID]map[int]struct{}
	//  是ipv4协议
	isIpV4 bool
	// 协议 icmp udp
	protocol string
	// 组装后的目标ip信息
	ipaddr *net.IPAddr
	// 输入的目标ip
	addr string
	// 生存时间
	TTL int
	// 日志记录
	Log *scanLog.LogConf
	//
	done chan interface{}
	lock sync.Mutex
}

func New(host string) *Metadata {
	loadLog, err := scanLog.LoadLog("Ping日志")
	r := rand.New(rand.NewSource(utils.GetSeed()))
	firstUUID := uuid.New()
	var firstSequence = map[uuid.UUID]map[int]struct{}{}
	firstSequence[firstUUID] = make(map[int]struct{})
	if err != nil {
		panic(err)
	}
	return &Metadata{
		Count:        -1,
		Interval:     time.Second,
		Size:         timeSliceLength + trackerLength,
		Timeout:      time.Duration(math.MaxInt64),
		Log:          loadLog,
		addr:         host,
		done:         make(chan interface{}),
		id:           r.Intn(math.MaxUint16),
		trackerUUIDs: []uuid.UUID{firstUUID},
		ipaddr:       nil,
		isIpV4:       true,
		protocol:     "icmp",
		TTL:          64,
	}
}

func (p *Metadata) Stop() {
	p.lock.Lock()
	defer p.lock.Unlock()
	open := true
	select {
	case _, open = <-p.done:
	}
	if open {
		close(p.done)
	}
}

func (p *Metadata) Resolve() error {
	if len(p.addr) == 0 {
		err := errors.New("目标ip为空")
		p.Log.Error.Println(err)
		return err
	}
	addr, err := net.ResolveIPAddr("ip", p.addr)
	if err != nil {
		p.Log.Error.Println(err)
		return err
	}
	p.ipaddr = addr
	return nil
}

func NewPingMetadata(host string) (*Metadata, error) {
	p := New(host)
	return p, p.Resolve()
}

func (p *Metadata) Run() error {
	return nil
}
