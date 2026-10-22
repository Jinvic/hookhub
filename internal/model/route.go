package model

type Route struct {
	Name    string   `yaml:"name"`
	Match   Match    `yaml:"match"`
	Actions []Action `yaml:"actions"`
}

type Match struct {
	Source string `yaml:"source"`
	Type   string `yaml:"type"`
}

type Action struct {
	Target   string   `yaml:"target"`
	Template Template `yaml:"template"`
}

type Template struct {
	Title    string            `yaml:"title"`
	Body     string            `yaml:"body"`
	Format   string            `yaml:"format"`
	Priority string            `yaml:"priority"`
	Extra    map[string]string `yaml:"extra"`
}
