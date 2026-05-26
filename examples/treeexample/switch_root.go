package treeexample

import "DecisionTreeDemoNew/tree"

type SwitchRoot struct {
	AbstractSupport
	accountNode *AccountNode
}

func NewSwitchRoot(accountNode *AccountNode) *SwitchRoot {
	return &SwitchRoot{AbstractSupport: NewAbstractSupport(), accountNode: accountNode}
}

func (n *SwitchRoot) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *SwitchRoot) DoApply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Route(n, req, ctx)
}

func (n *SwitchRoot) Get(_ string, _ *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	return n.accountNode, nil
}
