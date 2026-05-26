package tree

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"sort"
)

// NASNode NAS虚机选择节点（单独部署模式）
// 算法：读IOPS → CPU档位匹配 → 迭代计算节点数 n → 带宽校验 + 不均匀度校验
type NASNode struct {
	LeafNode
}

func NewNASNode(configJSON []byte) *NASNode {
	return &NASNode{
		LeafNode: LeafNode{
			VMBaseNode: VMBaseNode{
				NodeKey:    "nas",
				NodeName:   "NAS前端接入节点",
				NodeType:   "leaf",
				ConfigJSON: configJSON,
			},
		},
	}
}

func (n *NASNode) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.LeafNode.ApplyBase(n, req, ctx)
}

// DoApply 执行 NAS 节点核心逻辑
func (n *NASNode) DoApply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	// 1. 计算读 IOPS 需求
	readIOPS := int(float64(req.BandwidthMB) * n.Config.ReadIopsDensity)
	ctx.ReadIOPS = readIOPS

	// 2. 选择性能档位
	performanceTier := n.Config.NASPerformanceTierX86
	if len(performanceTier) == 0 {
		performanceTier = n.Config.NASPerformanceTierARM
	}
	if len(performanceTier) == 0 {
		// 默认档位
		performanceTier = map[string]int{
			"4":  5000,
			"8":  10000,
			"16": 16000,
		}
	}

	// 3. 找到满足读 IOPS 需求的最小 CPU
	minCPUByDemand := FindMinCPUForSetattr(performanceTier, readIOPS)
	minCPU := n.Config.MinCPU
	if minCPUByDemand > minCPU {
		minCPU = minCPUByDemand
	}
	if minCPU == 0 {
		minCPU = 4
	}

	// 4. 过滤满足起步要求的规格
	qualified := FilterSpecsByMinRequirements(ctx.CandidateSpecs, minCPU, n.Config.MinMemory)
	if len(qualified) == 0 {
		return nil, domain.ErrNoSpecFound
	}

	// 5. 对每种规格计算需要的节点数
	var results []domain.SelectionResult

	for _, spec := range qualified {
		nodeCount := n.CalculateNodeCount(req, spec)
		if nodeCount <= 0 {
			continue
		}

		results = append(results, domain.SelectionResult{
			NAS: &domain.VMQuantity{
				SpecName: spec.Name,
				Quantity: nodeCount,
				CPUCores: spec.CPUCores,
			},
			TotalCost: spec.CPUCores * nodeCount,
		})
	}

	// 6. 按成本排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCost < results[j].TotalCost
	})

	return results, nil
}

// CalculateNodeCount 计算 NAS 需要的节点数
func (n *NASNode) CalculateNodeCount(req domain.SelectionInput, spec domain.VMSpec) int {
	maxNodes := req.VipCount // n 的最大值为 VIP 数量
	if maxNodes == 0 {
		maxNodes = 10 // 默认保护
	}

	unevennessLimit := n.Config.UnevennessLimit
	if unevennessLimit == 0 {
		unevennessLimit = 0.3 // 默认30%
	}
	bandwidthReserved := n.Config.BandwidthReserved
	if bandwidthReserved == 0 {
		bandwidthReserved = 0.9 // 默认预留10%
	}

	// 从最小节点数开始迭代，直到满足带宽约束
	for nodeCount := 1; nodeCount <= maxNodes; nodeCount++ {
		// 1. 不均匀度校验
		unevenness := domain.CalcUnevenness(req.VipCount, nodeCount)
		if unevenness > unevennessLimit {
			continue
		}

		// 2. 带宽校验：bw <= 90% * bw_net
		vipPerNode := domain.CeilInt(req.VipCount, nodeCount)

		// 单虚机带宽
		vmNetBwMBps := spec.NetBwMBps()
		if vmNetBwMBps == 0 && spec.NetBwGbps > 0 {
			vmNetBwMBps = domain.GbpsToMBps(spec.NetBwGbps)
		}

		// 总网络带宽能力 = 单虚机带宽 × 虚机数量
		totalNetBw := float64(vmNetBwMBps) * float64(nodeCount)
		// 实际可用带宽（考虑不均匀分布）
		vipRatio := float64(vipPerNode) / float64(req.VipCount)
		availableBw := totalNetBw / vipRatio

		// 业务带宽需要 <= 90% * availableBw
		if float64(req.BandwidthMB) <= bandwidthReserved*availableBw {
			return nodeCount
		}
	}

	return maxNodes // 返回最大尝试值
}
