package treeexample

import "DecisionTreeDemoNew/tree"

type DefaultStrategyFactory struct {
	rootNode *RootNode
}

func NewDefaultStrategyFactory(rootNode *RootNode) *DefaultStrategyFactory {
	return &DefaultStrategyFactory{rootNode: rootNode}
}

func (f *DefaultStrategyFactory) StrategyHandler() tree.StrategyHandler[string, *DynamicContext, string] {
	return f.rootNode
}
