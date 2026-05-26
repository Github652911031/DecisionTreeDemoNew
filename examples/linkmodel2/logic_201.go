package linkmodel2

import "DecisionTreeDemoNew/link/model2"

type RuleLogic201 struct {
	model2.BaseHandler[string, *DynamicContext, XxxResponse]
}

func (r *RuleLogic201) Apply(req string, ctx *DynamicContext) (XxxResponse, error) {
	return r.Next(req, ctx)
}
