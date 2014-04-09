package agent

import (
	"github.com/cloudfoundry-incubator/dropsonde-agent/emitter"
	"errors"
	"fmt"
	"sync/atomic"
	"log"
	"net"
)

var UdpListeningPort = &port{42420}
var TcpListeningPort = &port{42421}

func Run(stopChan <-chan struct{}) (err error) {
	if emitter.DefaultEmitter == nil {
		return errors.New("Could not start agent. No default emitter provided.")
	}

	internalStopChan := make(chan struct{})
	dataChan := make(chan []byte)
	defer close(dataChan)
	udpErrChan := make(chan error)
	tcpErrChan := make(chan error)

	go func() {
		udpErrChan <- runListener(listenUdp, UdpListeningPort, dataChan, internalStopChan)
		close(udpErrChan)
	}()

	go func() {
		tcpErrChan <- runListener(listenTcp, TcpListeningPort, dataChan, internalStopChan)
		close(tcpErrChan)
	}()

	loop: for {
		select {
		case data := <-dataChan:
		   emitter.DefaultEmitter.Emit(data)
		case <-stopChan:
		  break loop
		case err = <-udpErrChan:
			break loop
		case err = <-tcpErrChan:
			break loop
		}
	}

	/*
		Cleaning up the various channels is a relatively delicate process
		as the following must occur in a specific order

		1. request TCP/UDP listeners stop
		2. drain the shared data channel as listeners may push messages before receiving
		   a close request
		3. drain the error channels for the TCP/UDP listeners until the listeners have stopped
	 */

	// request listeners to stop
	close(internalStopChan)

	// drain dataChan (ensures that listeners will be unblocked and can stop properly)
	go func() {
		for _ = range(dataChan) {}
	}()

	// wait for listeners to actually stop
	for _ = range udpErrChan {}
	for _ = range tcpErrChan {}

	return
}

type port struct { val int32 }

func (p *port) Set(val int) {
	atomic.StoreInt32(&p.val, int32(val))
}

func (p *port) Get() int {
	return int(atomic.LoadInt32(&p.val))
}

func runListener(listenPacket listenPacketFunc, port *port, dataChan chan<- []byte, stopChan <-chan struct{}) (err error) {
	conn, err := listenPacket(fmt.Sprintf(":%d", port.Get()))
	if err != nil {
		return
	}
	port.Set(conn.Port())

	go func() {
		<-stopChan
		conn.Close()
	}()

	for {
		buffer := make([]byte, 4096)
		var n int
		var addr net.Addr
		n, addr, err = conn.ReadFrom(buffer) // errors returned from here are used for control
		if err != nil {
			return
		}

		if n == 0 {
			log.Printf("Warning: zero bytes read from client: %v\n", addr.String())
		} else {
			dataChan <- buffer[0:n]
		}
	}
}
