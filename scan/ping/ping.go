package ping

import (
	"ScanTodo/scanLog"
	"errors"
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
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
	// 记录序列号
	awaitingSequences map[uuid.UUID]map[int]struct{}
	// 只有 ipv4
	network string
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
	return &Metadata{}
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
	addr, err := net.ResolveIPAddr(p.network, p.addr)
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
