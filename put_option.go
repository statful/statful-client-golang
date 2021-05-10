package statful

type putOptions struct {
	user string
}

type PutOption func(*putOptions)

func newPutOptions(opts []PutOption) *putOptions {

	p := &putOptions{}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithUser(user string) PutOption {
	return func(p *putOptions) {
		p.user = user
	}
}
