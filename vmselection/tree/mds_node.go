package tree

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"sort"
)

// MDSNode MDS虚机选择节点
// 算法：带宽 → 写IOPS → setattr需求 → CPU档位匹配 → 固定2台
type MDSNode struct {
	LeafNode
}

// NewMDSNode 创建MDS节点
func NewMDSNode(configJSON []byte) *MDSNode {
	return &MDSNode{
		LeafNode: LeafNode{
			VMBaseNode: VMBaseNode{
				NodeKey:    "mds",
				NodeName:   "MDS元数据服务节点",
				NodeType:   "leaf",
				ConfigJSON: configJSON,
			},
		},
	}
}

// Apply 实现 StrategyHandler 接口
func (n *MDSNode) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.LeafNode.ApplyBase(n, req, ctx)
}

// DoApply 执行 MDS 节点核心逻辑
func (n *MDSNode) DoApply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	// 1. 计算 setattr 需求
	writeIOPS := int(float64(req.BandwidthMB) * n.Config.WriteIopsDensity)
	setattrDemand := int(float64(writeIOPS) * n.Config.SetattrRatio)
	ctx.WriteIOPS = writeIOPS
	ctx.SetattrDemand = setattrDemand

	// 2. 选择性能档位
	performanceTier := n.Config.MdsPerformanceTier
	if len(performanceTier) == 0 {
		// 默认档位（兼容旧配置）
		performanceTier = map[string]int{
			"4":  25000,
			"8":  60000,
			"16": 90000,
		}
	}

	// 3. 找到满足 setattr 需求的最小 CPU
	minCPUByDemand := FindMinCPUForSetattr(performanceTier, setattrDemand)

	// 4. 取起步配置和需求计算的最大值
	minCPU := n.Config.MinCPU
	if minCPUByDemand > minCPU {
		minCPU = minCPUByDemand
	}
	if minCPU == 0 {
		minCPU = 4 // 默认起步
	}

	// 5. 过滤满足起步要求的规格
	qualified := FilterSpecsByMinRequirements(ctx.CandidateSpecs, minCPU, n.Config.MinMemory)
	if len(qualified) == 0 {
		return nil, domain.ErrNoSpecFound
	}

	// 6. MDS 固定 2 台
	mdsCount := 2
	if n.Config.MdsNodeCount > 0 {
		mdsCount = n.Config.MdsNodeCount
	}

	results := make([]domain.SelectionResult, 0, len(qualified))
	for _, spec := range qualified {
		results = append(results, domain.SelectionResult{
			MDS: domain.VMQuantity{
				SpecName: spec.Name,
				Quantity: mdsCount,
				CPUCores: spec.CPUCores,
			},
			TotalCost: spec.CPUCores * mdsCount,
		})
	}

	// 7. 按成本排序（低CPU优先）
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCost < results[j].TotalCost
	})

	return results, nil
}
