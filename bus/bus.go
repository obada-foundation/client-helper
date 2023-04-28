package bus

import (
	"time"

	"github.com/mustafaturan/bus/v3"
	"github.com/mustafaturan/monoton/v3"
	"github.com/mustafaturan/monoton/v3/sequencer"
)

// NewBus creating new bus instance
func NewBus() (*bus.Bus, error) {
	node := uint64(1)
	initialTime := uint64(time.Now().UnixNano())
	m, err := monoton.New(sequencer.NewMillisecond(), node, initialTime)
	if err != nil {
		return nil, err
	}

	// init an id generator
	var idGenerator bus.Next = m.Next

	// create a new bus instance
	b, err := bus.NewBus(idGenerator)
	if err != nil {
		return nil, err
	}

	return b, nil
}
