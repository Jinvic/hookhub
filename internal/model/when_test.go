package model

import "testing"

func TestCompileWhenEmpty(t *testing.T) {
	prg, err := CompileWhen("")
	if err != nil {
		t.Fatal(err)
	}
	if prg != nil {
		t.Fatal("expected nil program")
	}
	ok, err := EvalWhen(nil, &Event{}, nil)
	if err != nil || !ok {
		t.Fatalf("nil eval = %v %v", ok, err)
	}
}

func TestCompileWhenNotBool(t *testing.T) {
	_, err := CompileWhen(`"hello"`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompileWhenInvalid(t *testing.T) {
	_, err := CompileWhen("data.project.id ==")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEvalWhenProjectID(t *testing.T) {
	prg, err := CompileWhen(`data.project.id == 2`)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate JSON unmarshal: numbers as float64
	e := &Event{
		Source: "vikunja",
		Type:   "task.reminder.fired",
		Data: map[string]any{
			"project": map[string]any{"id": float64(2), "title": "Inbox"},
			"task":    map[string]any{"priority": float64(5)},
		},
	}
	ok, err := EvalWhen(prg, e, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true")
	}

	e.Data["project"].(map[string]any)["id"] = float64(3)
	ok, err = EvalWhen(prg, e, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected false")
	}
}

func TestEvalWhenNickname(t *testing.T) {
	prg, err := CompileWhen(`data.payload_raw.Comment.nickname != "bot"`)
	if err != nil {
		t.Fatal(err)
	}
	e := &Event{
		Data: map[string]any{
			"payload_raw": map[string]any{
				"Comment": map[string]any{"nickname": "alice"},
			},
		},
	}
	ok, err := EvalWhen(prg, e, nil)
	if err != nil || !ok {
		t.Fatalf("alice: %v %v", ok, err)
	}

	e.Data["payload_raw"].(map[string]any)["Comment"].(map[string]any)["nickname"] = "bot"
	ok, err = EvalWhen(prg, e, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected bot filtered")
	}
}

func TestEvalWhenUsesVars(t *testing.T) {
	prg, err := CompileWhen(`vars.timezone == "Asia/Shanghai"`)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := EvalWhen(prg, &Event{}, map[string]string{"timezone": "Asia/Shanghai"})
	if err != nil || !ok {
		t.Fatalf("%v %v", ok, err)
	}
}
