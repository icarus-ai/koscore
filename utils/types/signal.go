package types

import "sync/atomic"

type Channel chan uint8

type Signal struct {
	is  atomic.Bool
	sig Channel
}

func NewSignal() *Signal {
	return &Signal{sig: make(Channel)}
}

func (m *Signal) Reset() {
	close(m.sig)
	m.sig = make(Channel)
	m.is.Store(false)
}

func (m *Signal) Run() bool {
	if m.is.Load() {
		return true
	}
	m.is.Store(true)
	return false
}

func (m *Signal) Store(v bool)  { m.is.Store(v) }
func (m *Signal) Load() bool    { return m.is.Load() }
func (m *Signal) Case() Channel { return m.sig }
func (m *Signal) Signal()       { m.sig <- 0 }
func (m *Signal) StopAndWait() uint8 {
	if !m.is.Load() {
		return 0
	}
	m.sig <- 1
	return <-m.sig
}
