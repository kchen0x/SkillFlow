package main

type processRole string

const (
	processRoleDaemon processRole = "daemon"
	processRoleUI     processRole = "ui"

	internalDaemonFlag = "--internal-daemon"
	internalUIFlag     = "--internal-ui"
)

var activeProcessRole = processRoleDaemon

func determineProcessRole(args []string) (processRole, []string) {
	if len(args) == 0 {
		return processRoleDaemon, args
	}

	filtered := make([]string, 0, len(args))
	role := processRoleDaemon
	for _, arg := range args {
		switch arg {
		case internalUIFlag:
			role = processRoleUI
		case internalDaemonFlag:
			role = processRoleDaemon
		default:
			filtered = append(filtered, arg)
		}
	}
	return role, filtered
}
