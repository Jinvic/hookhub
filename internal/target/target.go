package target

import (
	"context"

	"github.com/Jinvic/hookhub/internal/model"
)

type Target interface {
	Name() string
	Send(ctx context.Context, msg *model.Message, cfg map[string]any) error
}
