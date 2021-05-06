package statful

type put struct {
	User string
}

type PutOption func(*put)

func NewPut(opts []PutOption) *put {

	const defaultUser = ""

	p := &put{
		User: defaultUser,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithUser(user string) PutOption {
	return func(p *put) {
		p.User = user
	}
}
