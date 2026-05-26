package model1

type LogicLink[T any, D any, R any] interface {
	Apply(req T, ctx D) (R, error)
	AppendNext(next LogicLink[T, D, R]) LogicLink[T, D, R]
	Next() LogicLink[T, D, R]
}

type BaseLogicLink[T any, D any, R any] struct {
	next LogicLink[T, D, R]
}

func (b *BaseLogicLink[T, D, R]) AppendNext(next LogicLink[T, D, R]) LogicLink[T, D, R] {
	b.next = next
	return next
}

func (b *BaseLogicLink[T, D, R]) Next() LogicLink[T, D, R] {
	return b.next
}

func (b *BaseLogicLink[T, D, R]) CallNext(req T, ctx D) (R, error) {
	if b.next == nil {
		var zero R
		return zero, ErrNextNotSet
	}
	return b.next.Apply(req, ctx)
}
