package model

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderContext is the data root for template execution.
type RenderContext struct {
	Event *Event
	Vars  map[string]string // per source-instance vars (e.g. base_url)
}

// Render builds a Message by executing Title/Body/Priority/Extra as text/templates
// against ctx (Event + Vars). Format is copied as-is. Missing map keys fail the render.
func (t Template) Render(ctx RenderContext) (Message, error) {
	if ctx.Event == nil {
		return Message{}, fmt.Errorf("render: nil event")
	}
	if ctx.Vars == nil {
		ctx.Vars = map[string]string{}
	}

	title, err := execTemplate("title", t.Title, ctx)
	if err != nil {
		return Message{}, fmt.Errorf("render title: %w", err)
	}
	body, err := execTemplate("body", t.Body, ctx)
	if err != nil {
		return Message{}, fmt.Errorf("render body: %w", err)
	}
	priority, err := execTemplate("priority", t.Priority, ctx)
	if err != nil {
		return Message{}, fmt.Errorf("render priority: %w", err)
	}

	var extra map[string]string
	if len(t.Extra) > 0 {
		extra = make(map[string]string, len(t.Extra))
		for k, v := range t.Extra {
			s, err := execTemplate("extra."+k, v, ctx)
			if err != nil {
				return Message{}, fmt.Errorf("render extra %q: %w", k, err)
			}
			extra[k] = s
		}
	}

	return Message{
		Title:    title,
		Body:     body,
		Format:   t.Format,
		Priority: priority,
		Extra:    extra,
	}, nil
}

func execTemplate(name, text string, data any) (string, error) {
	if text == "" {
		return "", nil
	}
	tmpl, err := template.New(name).Funcs(templateFuncs()).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
