package domain

import "errors"

var (
	ErrMainMemoryNotFound = errors.New("main memory not found")
	ErrModuleNotFound     = errors.New("module memory not found")
	ErrModuleExists       = errors.New("module memory already exists")
	ErrInvalidModuleName  = errors.New("invalid module memory name")
	ErrReadOnlyFile       = errors.New("memory file is read-only or locked")
	ErrNameCollision      = errors.New("sf- prefixed file already exists in agent rules directory")
)
