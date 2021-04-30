package statful

type Put struct {
	User string
}

type PutOption func(*Put)

func NewPut(opts []PutOption) *Put {

	const defaultUser = ""

	p := &Put{
		User: defaultUser,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithUser(user string) PutOption {
	return func(p *Put) {
		p.User = user
	}
}
