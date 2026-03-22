package domain

type PushMode string

const (
	PushModeMerge    PushMode = "merge"
	PushModeTakeover PushMode = "takeover"
)
