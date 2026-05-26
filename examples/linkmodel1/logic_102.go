package linkmodel1

import "DecisionTreeDemoNew/link/model1"

type RuleLogic102 struct {
	model1.BaseLogicLink[string, *DynamicContext, string]
}

func (r *RuleLogic102) Apply(_ string, _ *DynamicContext) (string, error) {
	return "link model01 单实例链", nil
}
