package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrParseInitMessage     = errors.New("failed to parse init message")
	ErrParseTransferMessage = errors.New("failed to parse transfer message")
	ErrUnknownMessageType   = errors.New("unknown message type")
	ErrDispatchFailed       = errors.New("dispatch failed")
)

type messageDispatch func(context.Context, []byte) error

type Dispatcher struct {
	stdin      io.Reader
	controller *Controller
}

func NewDispatcher(stdin io.Reader, controller *Controller) *Dispatcher {
	return &Dispatcher{stdin, controller}
}

func (d *Dispatcher) initMessage(_ context.Context, raw []byte) error {
	var msg Init
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: %w", ErrParseInitMessage, err)
	}

	return d.controller.init(msg)
}

func (d *Dispatcher) transferMessage(ctx context.Context, raw []byte) error {
	var msg Transfer
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("%w: %w", ErrParseTransferMessage, err)
	}

	return d.controller.transfer(ctx, msg)
}

func (d *Dispatcher) terminateMessage(_ context.Context, _ []byte) error {
	return io.EOF
}

func (d *Dispatcher) dispatch(ctx context.Context, msg []byte) error {
	dispatchers := map[string]messageDispatch{
		"init":      d.initMessage,
		"upload":    d.transferMessage,
		"download":  d.transferMessage,
		"terminate": d.terminateMessage,
	}

	var msgType struct {
		Event string `json:"event"`
	}

	if err := json.Unmarshal(msg, &msgType); err != nil {
		return fmt.Errorf("%w: %w", ErrUnknownMessageType, err)
	}

	parser, exists := dispatchers[msgType.Event]
	if !exists {
		return fmt.Errorf("%w: %s", ErrUnknownMessageType, msgType.Event)
	}

	return parser(ctx, msg)
}

func (d *Dispatcher) ListenAndServe(ctx context.Context) error {
	scanner := bufio.NewScanner(d.stdin)

	for scanner.Scan() {
		err := d.dispatch(ctx, scanner.Bytes())
		if err != nil && errors.Is(err, io.EOF) {
			return io.EOF
		} else if err != nil {
			return fmt.Errorf("%w: %w", ErrDispatchFailed, err)
		}
	}

	return nil
}
