package linkmodel2

import base "DecisionTreeDemoNew/link/model2"

type DynamicContext struct {
	*base.DynamicContext
	Age string
}

func NewDynamicContext() *DynamicContext {
	return &DynamicContext{DynamicContext: base.NewDynamicContext()}
}
