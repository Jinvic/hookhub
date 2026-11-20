package mox

func subjectIncoming(p *Incoming) string {
	if p.Subject != "" {
		return p.Subject
	}
	if len(p.From) > 0 && p.From[0].Address != "" {
		return p.From[0].Address
	}
	return ""
}

func subjectOutgoing(p *Outgoing) string {
	if p.Subject != "" {
		return p.Subject
	}
	return p.Event
}
