package internal

import (
	"context"
	"io"
)

type DialMessage interface {
	Marshal() ([]byte, error)
}

type Dial struct {
	stdout   io.Writer
	messages chan DialMessage
}

func NewDial(stdout io.Writer, messages chan DialMessage) *Dial {
	return &Dial{stdout, messages}
}

func (d Dial) ListenAndServe(ctx context.Context) {
	for {
		select {
		case msg := <-d.messages:
			marshalledMessage, err := msg.Marshal()
			if err != nil {
				panic(err)
			}

			_, err = d.stdout.Write(marshalledMessage)
			if err != nil {
				panic(err)
			}

			_, err = d.stdout.Write([]byte("\n"))
			if err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
