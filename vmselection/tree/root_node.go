package tree

import (
	basetree "DecisionTreeDemoNew/tree"
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
)

// RootNode 根节点 - 根据部署模式路由到2个分支
type RootNode struct {
	RouterNode

	SeparateAll *SeparateAllBranch
	SeparateNS  *SeparateNSBranch
}

func NewRootNode() *RootNode {
	return &RootNode{
		RouterNode: RouterNode{
			VMBaseNode: VMBaseNode{
				NodeKey:  "root",
				NodeName: "决策树根节点",
				NodeType: "router",
			},
		},
	}
}

func (n *RootNode) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.RouterNode.ApplyBase(n, req, ctx)
}

func (n *RootNode) Get(req domain.SelectionInput, ctx *context.VMContext) (basetree.StrategyHandler[domain.SelectionInput, *context.VMContext, []domain.SelectionResult], error) {
	switch req.DeployMode {
	case domain.DeployModeSeparateAll:
		return n.SeparateAll, nil
	case domain.DeployModeSeparateNS:
		return n.SeparateNS, nil
	default:
		return nil, domain.ErrInvalidDeployMode
	}
}

func (n *RootNode) DoApply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.RouterNode.DoApply(n, req, ctx)
}

type SeparateAllBranch struct {
	RouterNode
	MDSNode   *MDSNode
	NASNode   *NASNode
	SPACENode *SPACENode
}

func NewSeparateAllBranch() *SeparateAllBranch {
	return &SeparateAllBranch{
		RouterNode: RouterNode{VMBaseNode: VMBaseNode{NodeKey: "separate_all", NodeName: "三进程单独部署分支", NodeType: "router"}},
	}
}

func (n *SeparateAllBranch) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	if n.MDSNode == nil || n.NASNode == nil || n.SPACENode == nil {
		return nil, domain.ErrNodeNotInitialized
	}
	mdsResults, err := n.MDSNode.Apply(req, ctx)
	if err != nil {
		return nil, err
	}
	nasResults, err := n.NASNode.Apply(req, ctx)
	if err != nil {
		return nil, err
	}
	spaceResults, err := n.SPACENode.Apply(req, ctx)
	if err != nil {
		return nil, err
	}
	return n.combineResults(mdsResults, nasResults, spaceResults), nil
}

func (n *SeparateAllBranch) combineResults(mds, nas, space []domain.SelectionResult) []domain.SelectionResult {
	if len(mds) == 0 || len(nas) == 0 || len(space) == 0 {
		return nil
	}
	var results []domain.SelectionResult
	for i := 0; i < 3 && i < len(mds); i++ {
		for j := 0; j < 3 && j < len(nas); j++ {
			for k := 0; k < 3 && k < len(space); k++ {
				results = append(results, domain.SelectionResult{
					MDS:       mds[i].MDS,
					NAS:       nas[j].NAS,
					SPACE:     space[k].SPACE,
					TotalCost: mds[i].TotalCost + nas[j].TotalCost + space[k].TotalCost,
				})
			}
		}
	}
	SortResultsByCost(results)
	return results
}

type SeparateNSBranch struct {
	RouterNode
	MDSNode      *MDSNode
	NASSPACENode *NASSPACENode
}

func NewSeparateNSBranch() *SeparateNSBranch {
	return &SeparateNSBranch{
		RouterNode: RouterNode{VMBaseNode: VMBaseNode{NodeKey: "separate_ns", NodeName: "NAS/SPACE合部分支", NodeType: "router"}},
	}
}

func (n *SeparateNSBranch) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	if n.MDSNode == nil || n.NASSPACENode == nil {
		return nil, domain.ErrNodeNotInitialized
	}
	mdsResults, err := n.MDSNode.Apply(req, ctx)
	if err != nil {
		return nil, err
	}
	nsResults, err := n.NASSPACENode.Apply(req, ctx)
	if err != nil {
		return nil, err
	}
	return n.combineResults(mdsResults, nsResults), nil
}

func (n *SeparateNSBranch) combineResults(mds, ns []domain.SelectionResult) []domain.SelectionResult {
	if len(mds) == 0 || len(ns) == 0 {
		return nil
	}
	var results []domain.SelectionResult
	for i := 0; i < 3 && i < len(mds); i++ {
		for j := 0; j < 3 && j < len(ns); j++ {
			results = append(results, domain.SelectionResult{
				MDS:       mds[i].MDS,
				NASSPACE:  ns[j].NASSPACE,
				TotalCost: mds[i].TotalCost + ns[j].TotalCost,
			})
		}
	}
	SortResultsByCost(results)
	return results
}
