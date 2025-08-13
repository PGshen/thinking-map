package base

import "github.com/cloudwego/eino/compose"

// AgentOption is the common option type for various agent and multi-agent implementations.
// For options intended to use with underlying graph or components, use WithComposeOptions to specify.
// For options intended to use with particular agent/multi-agent implementations, use WrapImplSpecificOptFn to specify.
type AgentOption struct {
	implSpecificOptFn any
	composeOptions    []compose.Option
}

// GetComposeOptions returns all compose options from the given agent options.
func GetComposeOptions(opts ...AgentOption) []compose.Option {
	var result []compose.Option
	for _, opt := range opts {
		result = append(result, opt.composeOptions...)
	}

	return result
}

// WithComposeOptions returns an agent option that specifies compose options.
func WithComposeOptions(opts ...compose.Option) AgentOption {
	return AgentOption{
		composeOptions: opts,
	}
}

// WrapImplSpecificOptFn returns an agent option that specifies a function to modify the implementation-specific options.
func WrapImplSpecificOptFn[T any](optFn func(*T)) AgentOption {
	return AgentOption{
		implSpecificOptFn: optFn,
	}
}

// GetImplSpecificOptions returns the implementation-specific options from the given agent options.
func GetImplSpecificOptions[T any](base *T, opts ...AgentOption) *T {
	if base == nil {
		base = new(T)
	}

	for i := range opts {
		opt := opts[i]
		if opt.implSpecificOptFn != nil {
			optFn, ok := opt.implSpecificOptFn.(func(*T))
			if ok {
				optFn(base)
			}
		}
	}

	return base
}
