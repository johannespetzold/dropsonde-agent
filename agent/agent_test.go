package agent_test

import (
	"dropsonde-agent/agent"
	"dropsonde-agent/emitter"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"time"
)

type FakeEmitter struct {
	Events chan emitter.Event
}

func (fe *FakeEmitter) Emit(event emitter.Event) (err error) {
	fe.Events <- event
	return nil
}

var _ = Describe("Agent", func() {
	Describe("Run", func() {
		var (
			eventChan   chan emitter.Event
			fakeEmitter *FakeEmitter
			stopChan    chan struct{}
			errChan     chan error
		)

		BeforeEach(func() {
			agent.UdpListeningPort.Set(0)
			agent.TcpListeningPort.Set(0)

			eventChan = make(chan emitter.Event, 10)
			fakeEmitter = &FakeEmitter{eventChan}
			emitter.DefaultEmitter = fakeEmitter

			stopChan = make(chan struct{})
			errChan = make(chan error)
		})

		It("checks that there is a default emitter", func(done Done) {
			defer close(done)
			emitter.DefaultEmitter = nil
			err := agent.Run(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Could not start agent. No default emitter provided."))
		})

		It("listens for UDP packets and emits them", func(done Done) {
			defer close(done)

			go func() {
				defer close(errChan)
				err := agent.Run(stopChan)
				errChan <- err
			}()

			Eventually(agent.UdpListeningPort.Get).ShouldNot(BeZero())

			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", agent.UdpListeningPort.Get()))
			Expect(err).ToNot(HaveOccurred())

			conn, err := net.DialUDP("udp", nil, addr)
			Expect(err).ToNot(HaveOccurred())

			data := []byte("test-data")

			conn.Write(data)

			event := <-fakeEmitter.Events
			Expect(event).To(Equal(data))

			close(stopChan)
			err = <-errChan
			Expect(err).ToNot(HaveOccurred())
		})

		It("listens for TCP packets and emits them", func(done Done) {
			defer close(done)

			go func() {
				defer close(errChan)
				err := agent.Run(stopChan)
				errChan <- err
			}()

			Eventually(agent.TcpListeningPort.Get).ShouldNot(BeZero())

			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", agent.TcpListeningPort.Get()))
			Expect(err).ToNot(HaveOccurred())

			conn, err := net.DialTCP("tcp", nil, addr)
			Expect(err).ToNot(HaveOccurred())

			data := []byte("test-data")

			_, err = conn.Write(data)
			Expect(err).ToNot(HaveOccurred())

			err = conn.Close()
			Expect(err).ToNot(HaveOccurred())

			event := <-fakeEmitter.Events
			Expect(event).To(Equal(data))

			close(stopChan)
			err = <-errChan
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when a client sends an empty packet", func() {
			It("does not emit anything", func(done Done) {
				defer close(done)

				go func() {
					defer close(errChan)
					err := agent.Run(stopChan)
					errChan <- err
				}()

				Eventually(agent.UdpListeningPort.Get).ShouldNot(BeZero())

				addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", agent.UdpListeningPort.Get()))
				Expect(err).ToNot(HaveOccurred())

				conn, err := net.DialUDP("udp", nil, addr)
				Expect(err).ToNot(HaveOccurred())

				emptyData := make([]byte, 0)
				data := []byte("test-data")

				conn.Write(emptyData)
				conn.Write(data)

				event := <-fakeEmitter.Events
				Expect(event).To(Equal(data))

				close(stopChan)
				err = <-errChan
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when a tcp client does not close the connection", func() {
			It("still emits packets sent by subsequent clients", func(done Done) {
				defer close(done)

				agent.TcpReadTimeout = 50 * time.Millisecond

				go func() {
					defer close(errChan)
					err := agent.Run(stopChan)
					errChan <- err
				}()

				Eventually(agent.TcpListeningPort.Get).ShouldNot(BeZero())

				addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", agent.TcpListeningPort.Get()))

				firstConn, err := net.DialTCP("tcp", nil, addr)
				Expect(err).ToNot(HaveOccurred())

				secondConn, err := net.DialTCP("tcp", nil, addr)
				Expect(err).ToNot(HaveOccurred())

				data := []byte("test-data")

				_, err = secondConn.Write(data)
				Expect(err).ToNot(HaveOccurred())

				err = secondConn.Close()
				Expect(err).ToNot(HaveOccurred())

				event := <-fakeEmitter.Events
				Expect(event).To(Equal(data))

				firstConn.Close()

				close(stopChan)
				err = <-errChan
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
