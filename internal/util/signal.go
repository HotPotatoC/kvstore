package util

import (
	"os"
	"os/signal"
)

// WaitForSignals blocks the thread and wait for the
// provided signals (os.Signal)
func WaitForSignals(sig ...os.Signal) <-chan os.Signal {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, sig...)
	return signalChan
}
