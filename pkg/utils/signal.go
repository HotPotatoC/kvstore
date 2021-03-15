package utils

import (
	"os"
	"os/signal"
)

func WaitForSignals(sig ...os.Signal) <-chan os.Signal {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, sig...)
	return signalChan
}
