package treeexample

import "DecisionTreeDemoNew/tree"

type RootNode struct {
	AbstractSupport
	switchRoot           *SwitchRoot
	AfterCalled          bool
	AfterExceptionCalled bool
}

func NewRootNode(switchRoot *SwitchRoot) *RootNode {
	return &RootNode{AbstractSupport: NewAbstractSupport(), switchRoot: switchRoot}
}

func (n *RootNode) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *RootNode) DoApply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Route(n, req, ctx)
}

func (n *RootNode) Get(_ string, _ *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	return n.switchRoot, nil
}

func (n *RootNode) ApplyAfter(_ string, _ *DynamicContext, _ string) error {
	n.AfterCalled = true
	return nil
}

func (n *RootNode) ApplyAfterException(_ string, _ *DynamicContext, _ error) error {
	n.AfterExceptionCalled = true
	return nil
}
