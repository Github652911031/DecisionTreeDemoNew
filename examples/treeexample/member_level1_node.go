package treeexample

import (
	"DecisionTreeDemoNew/tree"
	"encoding/json"
)

type MemberLevel1Node struct {
	AbstractSupport
}

func NewMemberLevel1Node() *MemberLevel1Node {
	return &MemberLevel1Node{AbstractSupport: NewAbstractSupport()}
}

func (n *MemberLevel1Node) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *MemberLevel1Node) DoApply(_ string, ctx *DynamicContext) (string, error) {
	payload, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	return "level1" + string(payload), nil
}

func (n *MemberLevel1Node) Get(_ string, _ *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	return nil, nil
}
