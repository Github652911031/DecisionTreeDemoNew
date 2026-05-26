package tree

type DynamicContext struct {
	values map[string]any
}

func NewDynamicContext() DynamicContext {
	return DynamicContext{values: map[string]any{}}
}

func (c *DynamicContext) SetValue(key string, value any) {
	if c.values == nil {
		c.values = map[string]any{}
	}
	c.values[key] = value
}

func (c *DynamicContext) GetValue(key string) any {
	if c.values == nil {
		return nil
	}
	return c.values[key]
}
