//go:build !darwin

package main

func prepareHelperRuntime() error {
	return nil
}

func runHelperEventLoop(quitCh <-chan struct{}) error {
	<-quitCh
	return nil
}

func stopHelperEventLoop() {}
