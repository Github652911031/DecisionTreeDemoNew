package tree

type StrategyHandler[T any, D any, R any] interface {
	Apply(req T, ctx D) (R, error)
}

type StrategyHandlerFunc[T any, D any, R any] func(req T, ctx D) (R, error)

func (f StrategyHandlerFunc[T, D, R]) Apply(req T, ctx D) (R, error) {
	return f(req, ctx)
}

type StrategyMapper[T any, D any, R any] interface {
	Get(req T, ctx D) (StrategyHandler[T, D, R], error)
}

func DefaultHandler[T any, D any, R any]() StrategyHandler[T, D, R] {
	return StrategyHandlerFunc[T, D, R](func(_ T, _ D) (R, error) {
		var zero R
		return zero, nil
	})
}
