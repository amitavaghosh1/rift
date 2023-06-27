package commander

import (
	"context"
	"errors"
)

type Commander interface {
	Parse(ctx context.Context, cmdStr ...string) error
	Run(ctx context.Context) error
}

type NoopCommander struct{}

func (n NoopCommander) Parse(ctx context.Context, cmdSSr ...string) error {
	return ErrUnInitialized
}

func (n NoopCommander) Run(ctx context.Context) error {
	return ErrUnInitialized
}

var ErrUnInitialized = errors.New("command_uninitialized")
