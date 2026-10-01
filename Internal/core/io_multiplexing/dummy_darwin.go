//go:build darwin

package io_multiplexing

import "errors"

type DummyMultiplexer struct{}

func CreateIOMultiplexer() (IOMultiplexer, error) {
	return &DummyMultiplexer{}, errors.New("IO multiplexing is only supported on Linux")
}

func (m *DummyMultiplexer) Monitor(event Event) error {
	return nil
}

func (m *DummyMultiplexer) Wait() ([]Event, error) {
	return nil, errors.New("IO multiplexing is only supported on Linux")
}

func (m *DummyMultiplexer) Close() error {
	return nil
}
