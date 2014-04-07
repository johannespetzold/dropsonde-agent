package emitter

var DefaultEmitter Emitter

type Emitter interface {
	Emit(event Event) error
}

type Event interface {
}
