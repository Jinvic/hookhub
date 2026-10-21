package model

type Message struct {
	Title    string
	Body     string
	Priority string // 供 ntfy 等使用，可选
	Tags     []string
	Format   string // "text" | "markdown"
	Extra    map[string]string
}
