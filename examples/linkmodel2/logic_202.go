package linkmodel2

import "DecisionTreeDemoNew/link/model2"

type RuleLogic202 struct {
	model2.BaseHandler[string, *DynamicContext, XxxResponse]
}

func (r *RuleLogic202) ApplyBefore(req string, ctx *DynamicContext) (XxxResponse, error) {
	switch req {
	case "1":
		return r.JumpToNext(req, ctx, XxxResponse{Info: "00000"})
	case "2":
		return r.Stop(req, ctx, XxxResponse{Info: "applyBefore 拦截结果"})
	default:
		return r.Next(req, ctx)
	}
}

func (r *RuleLogic202) Apply(req string, ctx *DynamicContext) (XxxResponse, error) {
	return r.Next(req, ctx)
}
