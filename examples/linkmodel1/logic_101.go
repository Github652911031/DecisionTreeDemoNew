package linkmodel1

import "DecisionTreeDemoNew/link/model1"

type RuleLogic101 struct {
	model1.BaseLogicLink[string, *DynamicContext, string]
}

func (r *RuleLogic101) Apply(req string, ctx *DynamicContext) (string, error) {
	return r.CallNext(req, ctx)
}
