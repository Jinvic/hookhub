package source

import (
	"net/http"

	"github.com/Jinvic/hookhub/internal/model"
)

type Source interface {
	Name() string
	Verify(req *http.Request, body []byte) error
	Parse(req *http.Request, body []byte) (*model.Event, error)
}
