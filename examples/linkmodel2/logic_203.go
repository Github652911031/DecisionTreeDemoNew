package linkmodel2

import "DecisionTreeDemoNew/link/model2"

type RuleLogic203 struct {
	model2.BaseHandler[string, *DynamicContext, XxxResponse]
}

func (r *RuleLogic203) Apply(req string, ctx *DynamicContext) (XxxResponse, error) {
	return r.Stop(req, ctx, XxxResponse{Info: "hi 小傅哥！"})
}
