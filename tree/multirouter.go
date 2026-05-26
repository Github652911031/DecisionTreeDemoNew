package tree

const shortCircuitContextKey = "__tree_short_circuit__"

type shortCircuitContext interface {
	SetValue(key string, value any)
	GetValue(key string) any
}

type ApplyBeforeResult[R any] struct {
	Result R
	Stop   bool
}

type MultiThreadNode[T any, D any, R any] interface {
	StrategyMapper[T, D, R]
	ApplyBefore(req T, ctx D) (ApplyBeforeResult[R], error)
	MultiThread(req T, ctx D) error
	DoApply(req T, ctx D) (R, error)
	ApplyAfter(req T, ctx D, result R) error
	ApplyAfterException(req T, ctx D, err error) error
}

type MultiThreadRouterBase[T any, D any, R any] struct {
	RouterBase[T, D, R]
}

func NewMultiThreadRouterBase[T any, D any, R any]() MultiThreadRouterBase[T, D, R] {
	return MultiThreadRouterBase[T, D, R]{RouterBase: NewRouterBase[T, D, R]()}
}

func (b *MultiThreadRouterBase[T, D, R]) Apply(node MultiThreadNode[T, D, R], req T, ctx D) (R, error) {
	before, err := node.ApplyBefore(req, ctx)
	if err != nil {
		_ = node.ApplyAfterException(req, ctx, err)
		var zero R
		return zero, err
	}
	if before.Stop {
		if c, ok := any(ctx).(shortCircuitContext); ok {
			c.SetValue(shortCircuitContextKey, true)
		}
		return before.Result, nil
	}

	if err := node.MultiThread(req, ctx); err != nil {
		_ = node.ApplyAfterException(req, ctx, err)
		var zero R
		return zero, err
	}

	result, err := node.DoApply(req, ctx)
	if err != nil {
		_ = node.ApplyAfterException(req, ctx, err)
		var zero R
		return zero, err
	}
	if c, ok := any(ctx).(shortCircuitContext); ok && c.GetValue(shortCircuitContextKey) == true {
		return result, nil
	}

	if err := node.ApplyAfter(req, ctx, result); err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}
