package ntfy

import (
	"context"

	"github.com/Jinvic/hookhub/internal/model"
)

const NtfyTarget = "ntfy"

type Ntfy struct {
	Config *Config
}

func New(config *Config) *Ntfy {
	return &Ntfy{Config: config}
}

func (n *Ntfy) Name() string {
	return NtfyTarget
}

func (n *Ntfy) Send(ctx context.Context, msg *model.Message) error {
	return nil
}
