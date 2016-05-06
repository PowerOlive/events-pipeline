package events

import (
	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("events-pipeline")
)

// Event
type Key string
type Vals map[string]interface{}

type Event struct {
	Key  Key
	Vals Vals
}

func NewEvent(k Key, vals *Vals) *Event {
	return &Event{
		Key:  k,
		Vals: *vals,
	}
}

// Bolt
type Bolt interface {
}

// Wire
type Wire struct {
	senders   []Sender
	receivers []Receiver
	events    *chan *Event
}

// Sender
type Sender interface {
	Bolt
	LinkOutlet(*Wire)
	Send(*Event) error
}

type SenderBase struct {
	outlets []*Wire
}

func (s *SenderBase) LinkOutlet(wire *Wire) {
	s.outlets = append(s.outlets, wire)
}

func (s *SenderBase) Send(evt *Event) error {
	for _, w := range s.outlets {
		*w.events <- evt
	}
	return nil
}

// Receiver
type Receiver interface {
	Bolt
	LinkInlet(*Wire)
	Receive(*Event) error
}

type ReceiverBase struct {
	inlets []*Wire
}

func (s *ReceiverBase) LinkInlet(wire *Wire) {
	s.inlets = append(s.inlets, wire)
}

func (s *ReceiverBase) Receive(evt *Event) error {
	return nil
}

// Emitter
type Emitter interface {
	Sender
	Emit(Key, *Vals) error
}

type EmitterBase struct {
	id string
	SenderBase
}

func NewEmitter(ID string) *EmitterBase {
	return &EmitterBase{
		id: ID,
	}
}

func (e *EmitterBase) Emit(k Key, v *Vals) error {
	return e.SenderBase.Send(NewEvent(k, v))
}

func (e *EmitterBase) ID() string {
	return e.id
}

// Sink
type Sink interface {
	Receiver
}

// Processor
type Processor interface {
	Bolt

	// Sender
	LinkOutlet(*Wire)
	Send(*Event) error

	// Receiver
	LinkInlet(*Wire)
	Receive(*Event) error

	Process(*Event)
	Ack(*Event)
}

type ProcessorBase struct {
	SenderBase
	ReceiverBase
}
