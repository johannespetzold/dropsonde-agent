package agent

import (
	"dropsonde-agent/emitter"
	"errors"
	"fmt"
	"net"
	"sync"
)

var UdpIncomingPort = 42420
var TcpIncomingPort = 42421

var assignedUdpIncomingPort int
var assignedUdpIncomingPortLock sync.RWMutex

func Start(stopChan <-chan struct{}) (err error) {
	if emitter.DefaultEmitter == nil {
		return errors.New("Could not start agent. No default emitter provided.")
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UdpIncomingPort))
	if err != nil {
		return err
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	go func() {
		<-stopChan
		listener.Close()
	}()
	setAssignedPort(listener.LocalAddr().(*net.UDPAddr).Port)

	for {
		buffer := make([]byte, 4096)
		n, err := listener.Read(buffer)
		if err != nil {
			select {
			case <-stopChan:
				// err is most likely due to listener closed, not an actual error
				return nil
			default:
			}
			return err
		}
		emitter.DefaultEmitter.Emit(buffer[0:n])
	}
}

func setAssignedPort(port int) {
	assignedUdpIncomingPortLock.Lock()
	assignedUdpIncomingPort = port
	assignedUdpIncomingPortLock.Unlock()
}

func GetIncomingPort() int {
	assignedUdpIncomingPortLock.RLock()
	defer assignedUdpIncomingPortLock.RUnlock()
	return assignedUdpIncomingPort

}
