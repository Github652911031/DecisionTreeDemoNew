package treeexample

import (
	"DecisionTreeDemoNew/tree"
	"errors"
)

type AccountNode struct {
	AbstractSupport
	memberLevel1Node *MemberLevel1Node
	memberLevel2Node *MemberLevel2Node
	accountType01    func() string
	accountType02    func() string
	levelFn          func() int
	forceErr         bool
}

func NewAccountNode(memberLevel1Node *MemberLevel1Node, memberLevel2Node *MemberLevel2Node, accountType01 func() string, accountType02 func() string, levelFn func() int, forceErr bool) *AccountNode {
	return &AccountNode{
		AbstractSupport:  NewAbstractSupport(),
		memberLevel1Node: memberLevel1Node,
		memberLevel2Node: memberLevel2Node,
		accountType01:    accountType01,
		accountType02:    accountType02,
		levelFn:          levelFn,
		forceErr:         forceErr,
	}
}

func (n *AccountNode) Apply(req string, ctx *DynamicContext) (string, error) {
	return n.base.Apply(n, req, ctx)
}

func (n *AccountNode) ApplyBefore(req string, _ *DynamicContext) (tree.ApplyBeforeResult[string], error) {
	if req == "1" {
		return tree.ApplyBeforeResult[string]{Result: "xxx", Stop: true}, nil
	}
	return tree.ApplyBeforeResult[string]{}, nil
}

func (n *AccountNode) MultiThread(_ string, ctx *DynamicContext) error {
	ctx.SetValue("accountType01", n.accountType01())
	ctx.SetValue("accountType02", n.accountType02())
	return nil
}

func (n *AccountNode) DoApply(req string, ctx *DynamicContext) (string, error) {
	if n.forceErr {
		return "", errors.New("strconv.Atoi: parsing \"1xxx\": invalid syntax")
	}
	ctx.Level = n.levelFn()
	return n.base.Route(n, req, ctx)
}

func (n *AccountNode) Get(_ string, ctx *DynamicContext) (tree.StrategyHandler[string, *DynamicContext, string], error) {
	accountType01, _ := ctx.GetValue("accountType01").(string)
	accountType02, _ := ctx.GetValue("accountType02").(string)
	if accountType01 == "账户冻结" || accountType02 == "拦截" || ctx.Level == 1 {
		return n.memberLevel1Node, nil
	}
	return n.memberLevel2Node, nil
}
