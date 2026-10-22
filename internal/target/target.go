package target

import (
	"context"

	"github.com/Jinvic/hookhub/internal/model"
)

type Target interface {
	Driver() string
	Send(ctx context.Context, msg *model.Message) error
}
