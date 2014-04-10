package emitter

var DefaultEmitter Emitter

type Emitter interface {
	Emit(data []byte) error
}
