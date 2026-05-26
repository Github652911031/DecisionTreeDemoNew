package model2

type ContextWithFlow interface {
	SetProceed(bool)
	Proceed() bool
	SetJump(bool)
	Jump() bool
}

type Handler[T any, D ContextWithFlow, R any] interface {
	Apply(req T, ctx D) (R, error)
	ApplyBefore(req T, ctx D) (R, error)
	ApplyAfter(req T, ctx D, result R) error
	ApplyAfterException(req T, ctx D, err error) error
}

type BaseHandler[T any, D ContextWithFlow, R any] struct{}

func (b *BaseHandler[T, D, R]) Next(_ T, ctx D) (R, error) {
	ctx.SetJump(false)
	ctx.SetProceed(true)
	var zero R
	return zero, nil
}

func (b *BaseHandler[T, D, R]) Stop(_ T, ctx D, result R) (R, error) {
	ctx.SetJump(false)
	ctx.SetProceed(false)
	return result, nil
}

func (b *BaseHandler[T, D, R]) JumpToNext(_ T, ctx D, result R) (R, error) {
	ctx.SetJump(true)
	ctx.SetProceed(true)
	return result, nil
}

func (b *BaseHandler[T, D, R]) ApplyBefore(_ T, ctx D) (R, error) {
	ctx.SetJump(false)
	var zero R
	return zero, nil
}

func (b *BaseHandler[T, D, R]) ApplyAfter(_ T, _ D, _ R) error {
	return nil
}

func (b *BaseHandler[T, D, R]) ApplyAfterException(_ T, _ D, _ error) error {
	return nil
}
