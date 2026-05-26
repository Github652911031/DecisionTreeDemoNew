package treeexample

import "DecisionTreeDemoNew/tree"

type AbstractSupport struct {
	base tree.MultiThreadRouterBase[string, *DynamicContext, string]
}

func NewAbstractSupport() AbstractSupport {
	return AbstractSupport{base: tree.NewMultiThreadRouterBase[string, *DynamicContext, string]()}
}

func (s *AbstractSupport) ApplyBefore(_ string, _ *DynamicContext) (tree.ApplyBeforeResult[string], error) {
	return tree.ApplyBeforeResult[string]{}, nil
}

func (s *AbstractSupport) MultiThread(_ string, _ *DynamicContext) error {
	return nil
}

func (s *AbstractSupport) ApplyAfter(_ string, _ *DynamicContext, _ string) error {
	return nil
}

func (s *AbstractSupport) ApplyAfterException(_ string, _ *DynamicContext, _ error) error {
	return nil
}
