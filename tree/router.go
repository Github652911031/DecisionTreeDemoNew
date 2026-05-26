package tree

type RouterBase[T any, D any, R any] struct {
	Default StrategyHandler[T, D, R]
}

func NewRouterBase[T any, D any, R any]() RouterBase[T, D, R] {
	return RouterBase[T, D, R]{Default: DefaultHandler[T, D, R]()}
}

func (b *RouterBase[T, D, R]) Route(mapper StrategyMapper[T, D, R], req T, ctx D) (R, error) {
	handler, err := mapper.Get(req, ctx)
	if err != nil {
		var zero R
		return zero, err
	}
	if handler != nil {
		return handler.Apply(req, ctx)
	}
	return b.Default.Apply(req, ctx)
}
