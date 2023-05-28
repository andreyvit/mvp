package mvplive

import (
	"fmt"

	"github.com/andreyvit/mvp/flake"
)

type ChannelFamily struct {
	Name string
}

type Channel struct {
	Family *ChannelFamily
	Topic  string
}

func (ch Channel) String() string {
	return fmt.Sprintf("%s:%s", ch.Family.Name, ch.Topic)
}

type Envelope struct {
	EventID  uint64
	DedupKey string
}

type Msg struct {
	ID flake.ID
	Envelope
	Data []byte
}
