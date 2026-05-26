package treeexample

import (
	"DecisionTreeDemoNew/tree"
	"encoding/json"
)

type MemberLevel2Node struct {
	AbstractSupport
}

func NewMemberLevel2Node() *MemberLevel2Node {
	return &MemberLevel2Node{AbstractSupport: NewAbstractSupport()}
}

func (n *MemberLevel2Node) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *MemberLevel2Node) DoApply(_ string, ctx *DynamicContext) (string, error) {
	payload, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	return "level2" + string(payload), nil
}

func (n *MemberLevel2Node) Get(_ string, _ *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	return nil, nil
}
