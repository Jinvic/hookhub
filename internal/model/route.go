package model

type Route struct {
	Name    string   `yaml:"name"`
	Match   Match    `yaml:"match"`
	Actions []Action `yaml:"actions"`
}

type Match struct {
	Source string `yaml:"source"`
	Type   string `yaml:"type"`
	When   string `yaml:"when"` // optional CEL; empty = true
}

type Action struct {
	Target   string   `yaml:"target"`
	When     string   `yaml:"when"` // optional CEL; empty = true
	Template Template `yaml:"template"`
}

type Template struct {
	Title    string            `yaml:"title"`
	Body     string            `yaml:"body"`
	Format   string            `yaml:"format"`
	Priority string            `yaml:"priority"`
	Extra    map[string]string `yaml:"extra"`
}
