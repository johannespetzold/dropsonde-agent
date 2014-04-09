package emitter_test

import (
	"bytes"
	"code.google.com/p/gogoprotobuf/proto"
	"dropsonde-agent/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
)

var _ = Describe("LoggingEmitter", func() {
	Describe("Emit", func() {

		Context("with valid data", func() {
			It("logs emitted messages", func() {

				logWriter := new(bytes.Buffer)
				log.SetOutput(logWriter)

				emitter := emitter.NewLoggingEmitter()

				envelope := &events.Envelope{
					Origin:    events.NewOrigin("job-name", 42),
					EventType: events.Envelope_Heartbeat.Enum(),
					Heartbeat: events.NewHeartbeat(1, 2, 3),
				}
				data, err := proto.Marshal(envelope)
				Expect(err).ToNot(HaveOccurred())

				err = emitter.Emit(data)
				Expect(err).ToNot(HaveOccurred())

				loggedText := string(logWriter.Bytes())

				expectedText := proto.CompactTextString(envelope)
				Expect(loggedText).To(ContainSubstring(expectedText))
			})
		})

		Context("with invalid data", func() {
			It("returns an error", func() {
				emitter := emitter.NewLoggingEmitter()

				data := []byte{1, 2, 3}

				err := emitter.Emit(data)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
