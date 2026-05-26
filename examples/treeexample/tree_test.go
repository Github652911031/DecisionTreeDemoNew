package treeexample

import (
	"DecisionTreeDemoNew/tree"
	"errors"
	"testing"
)

func buildTree(forceErr bool, accountType01 func() string, accountType02 func() string, levelFn func() int) (*DefaultStrategyFactory, *RootNode) {
	level1 := NewMemberLevel1Node()
	level2 := NewMemberLevel2Node()
	account := NewAccountNode(level1, level2, accountType01, accountType02, levelFn, forceErr)
	switchRoot := NewSwitchRoot(account)
	root := NewRootNode(switchRoot)
	factory := NewDefaultStrategyFactory(root)
	return factory, root
}

type applyAfterErrorNode struct {
	AbstractSupport
	applyAfterCalled          bool
	applyAfterExceptionCalled bool
}

func newApplyAfterErrorNode() *applyAfterErrorNode {
	return &applyAfterErrorNode{AbstractSupport: NewAbstractSupport()}
}

func (n *applyAfterErrorNode) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *applyAfterErrorNode) DoApply(_ string, _ *DynamicContext) (string, error) {
	return "ok", nil
}

func (n *applyAfterErrorNode) Get(_ string, _ *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	return nil, nil
}

func (n *applyAfterErrorNode) ApplyAfter(_ string, _ *DynamicContext, _ string) error {
	n.applyAfterCalled = true
	return errors.New("applyAfter failed")
}

func (n *applyAfterErrorNode) ApplyAfterException(_ string, _ *DynamicContext, _ error) error {
	n.applyAfterExceptionCalled = true
	return nil
}

func TestTreeApplyBeforeShortCircuit(t *testing.T) {
	factory, root := buildTree(false, func() string { return "账户可用" }, func() string { return "已授信" }, func() int { return 0 })
	result, err := factory.StrategyHandler().Apply("1", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "xxx" {
		t.Fatalf("unexpected result: %s", result)
	}
	if root.AfterCalled {
		t.Fatalf("root applyAfter should not run on short circuit")
	}
}

func TestTreeExceptionFlow(t *testing.T) {
	factory, root := buildTree(true, func() string { return "账户可用" }, func() string { return "已授信" }, func() int { return 0 })
	_, err := factory.StrategyHandler().Apply("2", NewDynamicContext())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !root.AfterExceptionCalled {
		t.Fatalf("expected root applyAfterException to be called")
	}
}

func TestTreeRouteToLevel1ByAccountFreeze(t *testing.T) {
	factory, root := buildTree(false, func() string { return "账户冻结" }, func() string { return "已授信" }, func() int { return 0 })
	result, err := factory.StrategyHandler().Apply("2", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) < 6 || result[:6] != "level1" {
		t.Fatalf("unexpected result: %s", result)
	}
	if !root.AfterCalled {
		t.Fatalf("expected root applyAfter to be called on successful route")
	}
	if root.AfterExceptionCalled {
		t.Fatalf("root applyAfterException should not run on successful route")
	}
}

func TestTreeRouteToLevel2(t *testing.T) {
	factory, root := buildTree(false, func() string { return "账户可用" }, func() string { return "已授信" }, func() int { return 0 })
	result, err := factory.StrategyHandler().Apply("2", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) < 6 || result[:6] != "level2" {
		t.Fatalf("unexpected result: %s", result)
	}
	if !root.AfterCalled {
		t.Fatalf("expected root applyAfter to be called on successful route")
	}
	if root.AfterExceptionCalled {
		t.Fatalf("root applyAfterException should not run on successful route")
	}
}

func TestTreeApplyAfterErrorDoesNotTriggerApplyAfterException(t *testing.T) {
	node := newApplyAfterErrorNode()
	_, err := node.Apply("2", NewDynamicContext())
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "applyAfter failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !node.applyAfterCalled {
		t.Fatalf("expected applyAfter to be called")
	}
	if node.applyAfterExceptionCalled {
		t.Fatalf("applyAfterException should not run when applyAfter returns error")
	}
}
