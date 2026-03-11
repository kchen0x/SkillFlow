package main

type processRole string

const (
	processRoleHelper processRole = "helper"
	processRoleUI     processRole = "ui"

	internalUIFlag = "--internal-ui"
)

var activeProcessRole = processRoleHelper

func determineProcessRole(args []string) (processRole, []string) {
	if len(args) == 0 {
		return processRoleHelper, args
	}

	filtered := make([]string, 0, len(args))
	role := processRoleHelper
	for _, arg := range args {
		if arg == internalUIFlag {
			role = processRoleUI
			continue
		}
		filtered = append(filtered, arg)
	}
	return role, filtered
}
