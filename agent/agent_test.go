package agent_test

import (
	"dropsonde-agent/agent"
	"dropsonde-agent/emitter"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

type FakeEmitter struct {
	Events chan emitter.Event
}

func (fe *FakeEmitter) Emit(event emitter.Event) (err error) {
	fe.Events <- event
	return nil
}

var _ = Describe("Agent", func() {
	Describe("Start", func() {
		It("listens for UDP packages and emits them", func(done Done) {
			defer close(done)

			eventChan := make(chan emitter.Event, 10)
			fakeEmitter := &FakeEmitter{eventChan}
			emitter.DefaultEmitter = fakeEmitter

			agent.UdpIncomingPort = 0

			stopChan := make(chan struct{})
			errChan := make(chan error)

			go func(){
				defer close(errChan)
				err := agent.Start(stopChan)
				errChan <- err
			}()

			Eventually(func() int { return agent.GetIncomingPort() }).ShouldNot(BeZero())

			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", agent.GetIncomingPort()))
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
		It("checks that there is a default emitter", func(done Done) {
				defer close(done)
				emitter.DefaultEmitter = nil
				err := agent.Start(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Could not start agent. No default emitter provided."))
			})

		It("listens for TCP packages and emits them", func() {

		})
	})

})
