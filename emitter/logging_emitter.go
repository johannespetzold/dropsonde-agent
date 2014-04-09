package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"log"
)

type loggingEmitter struct {
}

func NewLoggingEmitter() Emitter {
	return new(loggingEmitter)
}

func (e *loggingEmitter) Emit(data []byte) (err error) {
	envelope := new(events.Envelope)
	err = proto.Unmarshal(data, envelope)
	if err != nil {
		return
	}

	log.Printf("Emitting %s\n", proto.CompactTextString(envelope))
	return
}
