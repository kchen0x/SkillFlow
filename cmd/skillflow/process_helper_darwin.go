//go:build darwin

package main

/*
void skillflow_prepare_application(void);
void skillflow_run_application(void);
void skillflow_stop_application(void);
*/
import "C"

import goruntime "runtime"

func prepareHelperRuntime() error {
	goruntime.LockOSThread()
	C.skillflow_prepare_application()
	return nil
}

func runHelperEventLoop(_ <-chan struct{}) error {
	C.skillflow_run_application()
	return nil
}

func stopHelperEventLoop() {
	C.skillflow_stop_application()
}
