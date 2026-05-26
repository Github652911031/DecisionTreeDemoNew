package treeexample

import base "DecisionTreeDemoNew/tree"

type DynamicContext struct {
	base.DynamicContext
	Level int `json:"level"`
}

func NewDynamicContext() *DynamicContext {
	return &DynamicContext{DynamicContext: base.NewDynamicContext()}
}
