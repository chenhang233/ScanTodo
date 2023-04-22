package ping

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"time"
)

type packetConn interface {
	Close() error
	ICMPRequestType() icmp.Type
	ReadFrom(b []byte) (n int, ttl int, src net.Addr, err error)
	SetFlagTTL() error
	SetReadDeadline(t time.Time) error
	WriteTo(b []byte, dst net.Addr) (int, error)
	SetTTL(ttl int)
}

type icmpConn struct {
	c   *icmp.PacketConn
	ttl int
}

func (c *icmpConn) Close() error {
	return c.c.Close()
}

func (c *icmpConn) SetTTL(ttl int) {
	c.ttl = ttl
}

func (c *icmpConn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *icmpConn) WriteTo(b []byte, destination net.Addr) (int, error) {
	if c.c.IPv4PacketConn() != nil {
		if err := c.c.IPv4PacketConn().SetTTL(c.ttl); err != nil {
			return 0, err
		}
	}
	return c.c.WriteTo(b, destination)
}

type icmpV4Conn struct {
	icmpConn
}

func (c *icmpV4Conn) SetFlagTTL() error {
	err := c.c.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
	return err
}

func (c *icmpV4Conn) ICMPRequestType() icmp.Type {
	return ipv4.ICMPTypeEcho
}

func (c *icmpV4Conn) ReadFrom(b []byte) (int, int, net.Addr, error) {
	ttl := -1
	n, cm, src, err := c.c.IPv4PacketConn().ReadFrom(b)
	if cm != nil {
		ttl = cm.TTL
	}
	return n, ttl, src, err
}

type icmpV6Conn struct {
	icmpConn
}

func (c *icmpV6Conn) SetFlagTTL() error {
	return nil
}

func (c *icmpV6Conn) ICMPRequestType() icmp.Type {
	return ipv6.ICMPTypeEchoRequest
}

func (c *icmpV6Conn) ReadFrom(b []byte) (int, int, net.Addr, error) {
	ttl := -1
	n, _, src, err := c.c.IPv6PacketConn().ReadFrom(b)
	return n, ttl, src, err
}
