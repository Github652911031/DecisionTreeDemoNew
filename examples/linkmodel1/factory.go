package linkmodel1

import "DecisionTreeDemoNew/link/model1"

type Factory struct {
	ruleLogic101 *RuleLogic101
	ruleLogic102 *RuleLogic102
}

func NewFactory() *Factory {
	return &Factory{
		ruleLogic101: &RuleLogic101{},
		ruleLogic102: &RuleLogic102{},
	}
}

func (f *Factory) OpenLogicLink() model1.LogicLink[string, *DynamicContext, string] {
	f.ruleLogic101.AppendNext(f.ruleLogic102)
	return f.ruleLogic101
}
