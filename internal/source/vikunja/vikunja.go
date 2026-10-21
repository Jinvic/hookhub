package vikunja

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jinvic/hookhub/internal/model"
)

const VikunjaSource = "vikunja"

type Vikunja struct {
	Config *Config
}

func New(config *Config) *Vikunja {
	return &Vikunja{Config: config}
}

func (v *Vikunja) Name() string {
	return VikunjaSource
}

func (v *Vikunja) Verify(req *http.Request, body []byte) error {
	if v.Config.AuthType == AuthNone {
		return nil
	}

	signature := req.Header.Get("X-Vikunja-Signature")

	mac := hmac.New(sha256.New, []byte(v.Config.SecretKey))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (v *Vikunja) Parse(_ *http.Request, body []byte) (*model.Event, error) {
	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshal body: %w", err)
	}

	data, err := p.ParseData()
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(body)
	return &model.Event{
		ID:      fmt.Sprintf("%s:%s:%x", VikunjaSource, p.EventName, sum[:8]),
		Source:  VikunjaSource,
		Type:    string(p.EventName),
		Time:    p.Time,
		Subject: subjectFor(p.EventName, data),
		Data:    data,
		Raw:     body,
	}, nil
}
