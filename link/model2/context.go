package model2

type DynamicContext struct {
	proceed bool
	jump    bool
	values  map[string]any
}

func NewDynamicContext() *DynamicContext {
	return &DynamicContext{
		proceed: true,
		jump:    false,
		values:  map[string]any{},
	}
}

func (c *DynamicContext) SetValue(key string, value any) {
	c.values[key] = value
}

func (c *DynamicContext) GetValue(key string) any {
	return c.values[key]
}

func (c *DynamicContext) Proceed() bool     { return c.proceed }
func (c *DynamicContext) SetProceed(v bool) { c.proceed = v }
func (c *DynamicContext) Jump() bool        { return c.jump }
func (c *DynamicContext) SetJump(v bool)    { c.jump = v }
