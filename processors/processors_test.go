package processors

import (
	"testing"
	"time"

	"github.com/getlantern/testify/assert"

	events "github.com/getlantern/events-pipeline"
)

// Null Sink
type NullSink struct {
	*events.SinkBase
}

func NewNullSink(id string) *NullSink {
	return &NullSink{
		SinkBase: events.NewSinkBase(id),
	}
}

// Override just to add logging
func (s *NullSink) Receive(evt *events.Event) error {
	log.Tracef("SINK ID %v received event: %v with: %v", s.ID(), evt.Key, evt.Vals)
	return s.SinkBase.Receive(evt)
}

// Callback Sink
type CallbackSink struct {
	*events.SinkBase
	callback func(e *events.Event)
}

func NewCallbackSink(id string, cb func(e *events.Event)) *CallbackSink {
	return &CallbackSink{
		SinkBase: events.NewSinkBase(id),
		callback: cb,
	}
}

func (c *CallbackSink) Receive(e *events.Event) error {
	c.callback(e)
	return c.SinkBase.Receive(e)
}

func TestAggregator(t *testing.T) {
	evs := make(chan *events.Event, 3)

	emitter := events.NewEmitterBase("test-emitter", nil)
	sink := NewCallbackSink("test-sink", func(e *events.Event) {
		log.Tracef("Entering callback!")
		evs <- e
	})

	aggregator := NewAggregator(
		"test-aggregator",
		nil,
		AggregationDirective{"Karma", "level", AggregatorIntRunningSum, RunningSumIdentity},
		AggregationDirective{"Happiness", "level", AggregatorFloat64MovingAverage, MovingAverageIdentity},
	)

	pipeline := events.NewPipeline(emitter)

	_, err := pipeline.Plug(emitter, aggregator)
	assert.Nil(t, err, "Should be nil")
	_, err = pipeline.Plug(aggregator, sink)
	assert.Nil(t, err, "Should be nil")

	pipeline.Run()

	// Test Running Sum
	emitter.Emit("Karma", &events.Vals{"level": 20})
	e := <-evs
	assert.Equal(t, 20, e.Vals["level"], "Should hold this value")

	emitter.Emit("Karma", &events.Vals{"level": 20})
	e = <-evs
	assert.Equal(t, 40, e.Vals["level"], "Should hold this value")

	emitter.Emit("Karma", &events.Vals{"level": 20})
	e = <-evs
	assert.Equal(t, 60, e.Vals["level"], "Should hold this value")

	// Test moving average
	// With these values, we shouldn't have floating point errors, but i
	emitter.Emit("Happiness", &events.Vals{"level": 250.5})
	e = <-evs
	assert.Equal(t, 250.5, e.Vals["level"], "Should hold this value")

	emitter.Emit("Happiness", &events.Vals{"level": 0.5})
	e = <-evs
	assert.Equal(t, 125.5, e.Vals["level"], "Should hold this value")

	emitter.Emit("Happiness", &events.Vals{"level": 300.0})
	emitter.Emit("Happiness", &events.Vals{"level": 400.0})
	e = <-evs
	e = <-evs
	assert.Equal(t, 237.75, e.Vals["level"], "Should hold this value")

	pipeline.Stop()
}

func TestIdentityProcessor(t *testing.T) {
	emitter := events.NewEmitterBase("test-emitter", nil)
	sink := NewNullSink("test-sink")
	dummy := NewIdentityProcessor("test-processor")
	pipeline := events.NewPipeline(emitter)
	_, err := pipeline.Plug(emitter, dummy)
	assert.Nil(t, err, "Should be nil")
	_, err = pipeline.Plug(dummy, sink)
	assert.Nil(t, err, "Should be nil")

	pipeline.Run()

	emitter.Emit("Key A", &events.Vals{})
	emitter.Emit("Key B", &events.Vals{})
	time.Sleep(time.Millisecond * 20)

	pipeline.Stop()
}
