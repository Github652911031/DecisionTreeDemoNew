package model2

import "errors"

var ErrDynamicProceed = errors.New("current item dynamic proceed is error")

type BusinessChain[T any, D ContextWithFlow, R any] struct {
	name     string
	handlers []Handler[T, D, R]
}

func NewBusinessChain[T any, D ContextWithFlow, R any](name string) *BusinessChain[T, D, R] {
	return &BusinessChain[T, D, R]{name: name, handlers: make([]Handler[T, D, R], 0)}
}

func (c *BusinessChain[T, D, R]) Name() string {
	return c.name
}

func (c *BusinessChain[T, D, R]) Add(handler Handler[T, D, R]) {
	c.handlers = append(c.handlers, handler)
}

func (c *BusinessChain[T, D, R]) Apply(req T, ctx D) (R, error) {
	for _, item := range c.handlers {
		applyBefore, err := item.ApplyBefore(req, ctx)
		if err != nil {
			_ = item.ApplyAfterException(req, ctx, err)
			var zero R
			return zero, err
		}
		if !ctx.Proceed() {
			if afterErr := item.ApplyAfter(req, ctx, applyBefore); afterErr != nil {
				var zero R
				return zero, afterErr
			}
			return applyBefore, nil
		}

		if ctx.Jump() {
			continue
		}

		applyResult, err := item.Apply(req, ctx)
		if err != nil {
			_ = item.ApplyAfterException(req, ctx, err)
			var zero R
			return zero, err
		}
		if !ctx.Proceed() {
			if afterErr := item.ApplyAfter(req, ctx, applyResult); afterErr != nil {
				var zero R
				return zero, afterErr
			}
			return applyResult, nil
		}
	}

	var zero R
	return zero, ErrDynamicProceed
}

func NewArmory[T any, D ContextWithFlow, R any](name string, handlers ...Handler[T, D, R]) *BusinessChain[T, D, R] {
	chain := NewBusinessChain[T, D, R](name)
	for _, handler := range handlers {
		chain.Add(handler)
	}
	return chain
}
