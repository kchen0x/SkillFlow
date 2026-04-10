package main

import daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"

var daemonInvokeServiceFn = func(method string, params any, result any) error {
	return daemonruntime.InvokeService(daemonServicePathFn(), method, params, result)
}

func shouldProxyAppMethodsToDaemon() bool {
	return activeProcessRole == processRoleUI
}

func (a *App) invokeDaemonService(method string, params any, result any) error {
	err := daemonInvokeServiceFn(method, params, result)
	if err != nil {
		a.logErrorf("daemon service invoke failed: method=%s err=%v", method, err)
	}
	return err
}
