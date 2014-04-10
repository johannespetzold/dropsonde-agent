package agent

import (
	"io"
	"log"
	"net"
	"time"
)

var TcpReadTimeout = 5 * time.Second

type listenPacketFunc func(addr string) (conn PacketConn, err error)

type PacketConn interface {
	Port() int
	ReadFrom([]byte) (int, net.Addr, error) // Errors returned here should be fatal (i.e. meant to shut down the listener)
	Close() error
}

type tcpPacketConn struct {
	*net.TCPListener
}

func listenTcp(addr string) (conn PacketConn, err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	conn = &tcpPacketConn{listener.(*net.TCPListener)}
	return
}

func (c *tcpPacketConn) Port() int {
	return c.Addr().(*net.TCPAddr).Port
}

func (c *tcpPacketConn) ReadFrom(buffer []byte) (n int, addr net.Addr, err error) {
	conn, err := c.Accept()
	if err != nil {
		return
	}

	addr = conn.RemoteAddr()

	conn.SetReadDeadline(time.Now().Add(TcpReadTimeout))
	n, err = io.ReadFull(conn, buffer)

	/*
		Assume that errors returned from Accept() should affect the caller
		but read errors are not fatal
	*/
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			log.Printf("error while reading from TCP connection: %v\n", err)
		}
		err = nil
	}

	return
}

type udpConn struct {
	*net.UDPConn
}

func listenUdp(addr string) (conn PacketConn, err error) {
	netConn, err := net.ListenPacket("udp", addr)
	conn = &udpConn{netConn.(*net.UDPConn)}
	return
}

func (c *udpConn) Port() int {
	return c.LocalAddr().(*net.UDPAddr).Port
}
