package model

// Message is the channel-agnostic payload delivered to a Target.
// Target-specific options (e.g. ntfy tags/click) belong in Extra.
type Message struct {
	Title    string
	Body     string
	Format   string // "text" | "markdown"
	Priority string // optional urgency hint; targets map or ignore
	Extra    map[string]string
}
