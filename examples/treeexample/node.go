package treeexample

import "DecisionTreeDemoNew/tree"

type StrategyNode interface {
	tree.StrategyHandler[string, *DynamicContext, string]
	tree.StrategyMapper[string, *DynamicContext, string]
}
