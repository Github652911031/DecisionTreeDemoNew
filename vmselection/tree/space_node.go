package tree

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"sort"
)

// SPACENode SPACE数据存储节点（单独部署模式）
// 算法：读IOPS → CPU档位匹配 → 迭代计算节点数 n → 盘分布带宽校验 + 不均匀度校验
type SPACENode struct {
	LeafNode
}

func NewSPACENode(configJSON []byte) *SPACENode {
	return &SPACENode{
		LeafNode: LeafNode{
			VMBaseNode: VMBaseNode{
				NodeKey:    "space",
				NodeName:   "SPACE数据存储节点",
				NodeType:   "leaf",
				ConfigJSON: configJSON,
			},
		},
	}
}

func (n *SPACENode) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.LeafNode.ApplyBase(n, req, ctx)
}

// DoApply 执行 SPACE 节点核心逻辑
func (n *SPACENode) DoApply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	// 1. 计算读 IOPS 需求
	readIOPS := int(float64(req.BandwidthMB) * n.Config.ReadIopsDensity)
	ctx.ReadIOPS = readIOPS

	// 2. 选择性能档位
	performanceTier := n.Config.SPACEPerformanceTierX86
	if len(performanceTier) == 0 {
		performanceTier = n.Config.SPACEPerformanceTierARM
	}
	if len(performanceTier) == 0 {
		// 默认档位
		performanceTier = map[string]int{
			"4":  6000,
			"8":  12000,
			"16": 20000,
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
			SPACE: &domain.VMQuantity{
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

// CalculateNodeCount 计算 SPACE 需要的节点数
func (n *SPACENode) CalculateNodeCount(req domain.SelectionInput, spec domain.VMSpec) int {
	// n 的最大值为 EVS 云盘数量的一半
	maxNodes := req.DiskCount / 2
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
		unevenness := domain.CalcUnevenness(req.DiskCount, nodeCount)
		if unevenness > unevennessLimit {
			continue
		}

		// 2. 磁盘带宽校验：bw <= 90% * bw_disk
		disksPerNode := domain.CeilInt(req.DiskCount, nodeCount)

		// 单虚机磁盘带宽
		vmDiskBwMBps := spec.DiskBwMBps()
		if vmDiskBwMBps == 0 && spec.DiskBwGbps > 0 {
			vmDiskBwMBps = domain.GbpsToMBps(spec.DiskBwGbps)
		}

		// 总磁盘带宽能力
		totalDiskBw := float64(vmDiskBwMBps) * float64(nodeCount)
		// 实际可用带宽（考虑盘的不均匀分布）
		diskRatio := float64(disksPerNode) / float64(req.DiskCount)
		availableBw := totalDiskBw / diskRatio

		// 业务带宽需要 <= 90% * availableBw
		if float64(req.BandwidthMB) <= bandwidthReserved*availableBw {
			return nodeCount
		}
	}

	return maxNodes // 返回最大尝试值
}
