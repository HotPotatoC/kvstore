package client

import (
	"sync"
	"testing"
)

func TestSetBusyPreservesOptions(t *testing.T) {
	c := &Client{}
	c.AddFlag(FlagReadOnly | FlagCloseASAP)
	for _, busy := range []bool{false, true, false, true, false} {
		c.SetBusy(busy)
		flags := Flags(c.flags.Load())
		state := FlagNone
		if busy {
			state = FlagBusy
		}
		if flags&(FlagNone|FlagBusy) != state || flags&(FlagReadOnly|FlagCloseASAP) != (FlagReadOnly|FlagCloseASAP) {
			t.Fatalf("busy=%t flags=%b", busy, flags)
		}
	}
}

func TestSetBusyConcurrentCloseRequest(t *testing.T) {
	c := &Client{}
	c.SetBusy(false)
	c.AddFlag(FlagReadOnly)
	started, done := make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(started)
		for i := 0; i < 10000; i++ {
			c.SetBusy(true)
			c.SetBusy(false)
		}
		close(done)
	}()
	<-started
	c.AddFlag(FlagCloseASAP)
	for {
		flags := Flags(c.flags.Load())
		state := flags & (FlagNone | FlagBusy)
		if state != FlagNone && state != FlagBusy {
			t.Errorf("command state has missing or overlapping flags: %b", flags)
			break
		}
		if flags&(FlagReadOnly|FlagCloseASAP) != (FlagReadOnly | FlagCloseASAP) {
			t.Errorf("command transition lost options: %b", flags)
			break
		}
		select {
		case <-done:
			wg.Wait()
			return
		default:
		}
	}
	wg.Wait()
}
