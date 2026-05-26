package linkmodel2

import "DecisionTreeDemoNew/link/model2"

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Demo01() *model2.BusinessChain[string, *DynamicContext, XxxResponse] {
	return model2.NewArmory[string, *DynamicContext, XxxResponse](
		"demo01",
		&RuleLogic201{},
		&RuleLogic202{},
		&RuleLogic203{},
	)
}

func (f *Factory) Demo02() *model2.BusinessChain[string, *DynamicContext, XxxResponse] {
	return model2.NewArmory[string, *DynamicContext, XxxResponse](
		"demo02",
		&RuleLogic202{},
		&RuleLogic203{},
	)
}
