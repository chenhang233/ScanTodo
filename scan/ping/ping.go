package ping

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"bytes"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sync/errgroup"
	"math"
	"math/rand"
	"net"
	"sync"
	"syscall"
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

// packet 接收到的数据包
type packet struct {
	bytes   []byte
	byteLen int
	ttl     int
}

type Packet struct {
	// 当前包往返时间统计
	RTTs time.Duration
	// 目标主机
	IPAddr *net.IPAddr
	// 目标主机字符串
	Addr string
	// 数据包长度
	ByteLen int
	// 序列号
	Sequence int
	TTL      int
	// 当前标识符
	Identifier int
}

// Statistics 统计
type Statistics struct {
	// 已经收到包数量
	PacketsReceive int
	// 已经发送包数量
	PacketsSent int
	//收到重复包数量
	PacketsReceiveDuplicates int
	// 丢包率
	PacketLoss float64
	// 目标主机
	IPAddr *net.IPAddr
	// 目标主机字符串
	Addr string
	// 所有往返时间统计
	RTTs []time.Duration
	// 往返时间统计
	minRoundTripTime     time.Duration
	maxRoundTripTime     time.Duration
	averageRoundTripTime time.Duration
}

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
	// Identifier
	id       int
	sequence int
	// 记录序列号
	awaitingSequences map[uuid.UUID]map[int]struct{}
	//  是ipv4协议
	isIpV4 bool
	// 协议 icmp udp
	Protocol string
	// 组装后的目标ip信息
	Ipaddr *net.IPAddr
	// 输入的目标ip
	Addr string
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
		Count:             24,
		Interval:          time.Second,
		Size:              timeSliceLength + trackerLength,
		Timeout:           time.Duration(math.MaxInt64),
		Log:               loadLog,
		Addr:              host,
		done:              make(chan interface{}),
		id:                r.Intn(math.MaxUint16),
		trackerUUIDs:      []uuid.UUID{firstUUID},
		awaitingSequences: firstSequence,
		Ipaddr:            nil,
		isIpV4:            true,
		Protocol:          "icmp",
		TTL:               64,
		Source:            "0.0.0.0",
	}
}

func (p *Metadata) Stop() {
	p.Log.Info.Printf("Stop 调用: ")
	p.lock.Lock()
	defer p.lock.Unlock()

	open := true
	select {
	case _, open = <-p.done:
	default:
	}
	if open {
		close(p.done)
	}
}

func (p *Metadata) Resolve() error {
	if len(p.Addr) == 0 {
		err := errors.New("目标ip为空")
		p.Log.Error.Println(err)
		return err
	}
	addr, err := net.ResolveIPAddr("ip", p.Addr)
	if err != nil {
		p.Log.Error.Println(err)
		return err
	}
	p.Ipaddr = addr
	return nil
}

func NewPingMetadata(host string) (*Metadata, error) {
	p := New(host)
	return p, p.Resolve()
}

func (p *Metadata) Run() error {
	var conn packetConn
	var err error
	if p.Size < timeSliceLength+trackerLength {
		p.Log.Error.Println("数据包小于必须大小 \n")
		return nil
	}
	conn, err = p.listen()
	if err != nil {
		p.Log.Error.Println("监听错误: ", err)
		return err
	}
	defer conn.Close()
	conn.SetTTL(p.TTL)
	return p.run(conn)
}

func (p *Metadata) listen() (packetConn, error) {
	var conn packetConn
	var err error
	if p.isIpV4 {
		c := &icmpV4Conn{}
		c.c, err = icmp.ListenPacket(ipv4Proto[p.Protocol], p.Source)
		conn = c
	} else {
		conn = &icmpV6Conn{}
		c := &icmpV6Conn{}
		c.c, err = icmp.ListenPacket(ipv6Proto[p.Protocol], p.Source)
		conn = c
	}
	if err != nil {
		p.Log.Error.Println("listen: ", err)
		p.Stop()
		return nil, err
	}
	return conn, nil
}

func (p *Metadata) run(conn packetConn) error {
	err := conn.SetFlagTTL()
	if err != nil {
		p.Log.Warn.Printf(err.Error())
	}
	defer p.finish()
	receive := make(chan *packet, 5)
	setup := p.OnSetup
	if setup != nil {
		setup()
	}
	var g errgroup.Group
	g.Go(func() error {
		defer p.Stop()
		return p.receiveICMP(conn, receive)
	})
	g.Go(func() error {
		defer p.Stop()
		return p.MainLoop(conn, receive)
	})
	err = g.Wait()
	p.Log.Error.Printf("结束中:", err)
	return err
}

func (p *Metadata) MainLoop(conn packetConn, re <-chan *packet) error {
	timeout := time.NewTicker(p.Timeout)
	interval := time.NewTicker(p.Interval)
	defer func() {
		timeout.Stop()
		interval.Stop()
	}()
	err := p.sendICMP(conn)
	if err != nil {
		p.Log.Warn.Printf("第一次发包出现异常: ", err)
	}
	for {
		select {
		case <-p.done:
			p.Log.Warn.Printf("收到结束信号")
			return nil
		case <-timeout.C:
			p.Log.Warn.Printf("收到超时信号")
			return nil
		case <-interval.C:
			p.Log.Debug.Printf("收到间隔时间信号")
			if (p.Count > 0 && p.PacketsSent >= p.Count) || p.Count < 0 {
				p.Stop()
				continue
			}
			err := p.sendICMP(conn)
			if err != nil {
				p.Log.Warn.Printf("发包异常", err)
			}
		case r := <-re:
			err := p.processPacket(r)
			if err != nil {
				p.Log.Warn.Printf("处理收到的包异常", err)
			}
		}
		if (p.Count > 0 && p.Count <= p.PacketsSent) || p.Count < 0 {
			return nil
		}
	}
}

func (p *Metadata) receiveICMP(conn packetConn, re chan<- *packet) error {
	for {
		select {
		case <-p.done:
			return nil
		default:
			var n, ttl int
			var err error
			b := make([]byte, p.getMessageLength())
			err = conn.SetReadDeadline(time.Now().Add(p.Timeout))
			if err != nil {
				p.Log.Warn.Printf(err.Error())
			}
			n, ttl, _, err = conn.ReadFrom(b)
			if err != nil {
				p.Log.Error.Printf(err.Error())
				return err
			}
			select {
			case <-p.done:
				return nil
			case re <- &packet{bytes: b, byteLen: n, ttl: ttl}:
			}
		}
	}
}

func (p *Metadata) sendICMP(conn packetConn) error {
	curUUID := p.trackerUUIDs[len(p.trackerUUIDs)-1]
	encode, err := curUUID.MarshalBinary()
	if err != nil {
		p.Log.Error.Println(err)
		return err
	}
	t := append(utils.TimeToBytes(time.Now()), encode...)
	remainSize := p.Size - timeSliceLength - trackerLength
	if remainSize > 0 {
		t = append(t, bytes.Repeat([]byte{1}, remainSize)...)
	}
	body := &icmp.Echo{
		ID:   p.id,
		Seq:  p.sequence,
		Data: t,
	}
	msg := &icmp.Message{
		Type: conn.ICMPRequestType(),
		Code: 0,
		Body: body,
	}
	byteData, err := msg.Marshal(nil)
	if err != nil {
		p.Log.Error.Println(err)
	}
	for {
		_, err := conn.WriteTo(byteData, p.Ipaddr)
		if err != nil {
			opError, ok := err.(*net.OpError)
			if ok && opError.Err == syscall.ENOBUFS {
				p.Log.Error.Println("发包错误定位: ", err)
				continue
			}
			p.Log.Error.Println("发包错误日志: ", err)
			return err
		}
		handle := p.OnSend
		if handle != nil {
			packet := &Packet{
				ByteLen:    len(byteData),
				IPAddr:     p.Ipaddr,
				Addr:       p.Addr,
				Sequence:   p.sequence,
				Identifier: p.id,
			}
			handle(packet)
		}
		p.awaitingSequences[curUUID][p.sequence] = struct{}{}
		p.sequence++
		p.PacketsSent++
		if p.sequence > 65535 {
			u := uuid.New()
			p.trackerUUIDs = append(p.trackerUUIDs, u)
			p.awaitingSequences[u] = make(map[int]struct{})
			p.sequence = 0
		}
		break
	}
	return nil
}

func (p *Metadata) processPacket(receive *packet) error {
	return nil
}

func (p *Metadata) finish() {
	handle := p.OnFinish
	if handle != nil {
		handle(p.Statistics())
	}
}

func (p *Metadata) Statistics() *Statistics {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	sent := p.PacketsSent
	loss := float64(sent-p.PacketsReceive) / float64(sent*100)
	s := &Statistics{
		PacketsSent:              sent,
		PacketsReceive:           p.PacketsReceive,
		PacketsReceiveDuplicates: p.PacketsReceiveDuplicates,
		PacketLoss:               loss,
		IPAddr:                   p.Ipaddr,
		Addr:                     p.Addr,
		RTTs:                     p.RTTs,
		minRoundTripTime:         p.minRoundTripTime,
		maxRoundTripTime:         p.maxRoundTripTime,
		averageRoundTripTime:     p.averageRoundTripTime,
	}
	return s
}

func (p *Metadata) getMessageLength() int {
	if p.isIpV4 {
		return p.Size + 8 + ipv4.HeaderLen
	}
	return p.Size + 8 + ipv6.HeaderLen
}
