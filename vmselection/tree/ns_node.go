package tree

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"sort"
)

// NASSPACENode NAS/SPACE合部节点
// 同时考虑 VIP分布 和 盘分布 的带宽约束
type NASSPACENode struct {
	LeafNode
}

func NewNASSPACENode(configJSON []byte) *NASSPACENode {
	return &NASSPACENode{
		LeafNode: LeafNode{
			VMBaseNode: VMBaseNode{
				NodeKey:    "ns",
				NodeName:   "NAS/SPACE合部节点",
				NodeType:   "leaf",
				ConfigJSON: configJSON,
			},
		},
	}
}

func (n *NASSPACENode) Apply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	return n.LeafNode.ApplyBase(n, req, ctx)
}

func (n *NASSPACENode) DoApply(req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	readIOPS := int(float64(req.BandwidthMB) * n.Config.ReadIopsDensity)
	ctx.ReadIOPS = readIOPS

	performanceTier := n.Config.NSPerformanceTierX86
	if len(performanceTier) == 0 {
		performanceTier = n.Config.NSPerformanceTierARM
	}
	if len(performanceTier) == 0 {
		performanceTier = map[string]int{"4": 3000, "8": 8000, "16": 15000}
	}

	minCPUByDemand := FindMinCPUForSetattr(performanceTier, readIOPS)
	minCPU := n.Config.MinCPU
	if minCPUByDemand > minCPU {
		minCPU = minCPUByDemand
	}
	if minCPU == 0 {
		minCPU = 4
	}

	qualified := FilterSpecsByMinRequirements(ctx.CandidateSpecs, minCPU, n.Config.MinMemory)
	if len(qualified) == 0 {
		return nil, domain.ErrNoSpecFound
	}

	var results []domain.SelectionResult
	for _, spec := range qualified {
		nodeCount := n.CalculateNodeCount(req, spec)
		if nodeCount <= 0 {
			continue
		}
		results = append(results, domain.SelectionResult{
			NASSPACE: &domain.VMQuantity{
				SpecName: spec.Name,
				Quantity: nodeCount,
				CPUCores: spec.CPUCores,
			},
			TotalCost: spec.CPUCores * nodeCount,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].TotalCost < results[j].TotalCost })
	return results, nil
}

func (n *NASSPACENode) CalculateNodeCount(req domain.SelectionInput, spec domain.VMSpec) int {
	maxNodes := req.DiskCount / 2
	if maxNodes == 0 {
		maxNodes = 10
	}

	unevennessLimit := n.Config.UnevennessLimit
	if unevennessLimit == 0 {
		unevennessLimit = 0.3
	}
	bandwidthReserved := n.Config.BandwidthReserved
	if bandwidthReserved == 0 {
		bandwidthReserved = 0.9
	}

	for nodeCount := 1; nodeCount <= maxNodes; nodeCount++ {
		vipUnevenness := domain.CalcUnevenness(req.VipCount, nodeCount)
		if vipUnevenness > unevennessLimit {
			continue
		}
		diskUnevenness := domain.CalcUnevenness(req.DiskCount, nodeCount)
		if diskUnevenness > unevennessLimit {
			continue
		}

		vipPerNode := domain.CeilInt(req.VipCount, nodeCount)
		disksPerNode := domain.CeilInt(req.DiskCount, nodeCount)

		vmNetBwMBps := spec.NetBwMBps()
		vmDiskBwMBps := spec.DiskBwMBps()

		totalNetBw := float64(vmNetBwMBps) * float64(nodeCount)
		vipRatio := float64(vipPerNode) / float64(req.VipCount)
		availableNetBw := totalNetBw / vipRatio

		totalDiskBw := float64(vmDiskBwMBps) * float64(nodeCount)
		diskRatio := float64(disksPerNode) / float64(req.DiskCount)
		availableDiskBw := totalDiskBw / diskRatio

		availableBw := availableNetBw
		if availableDiskBw < availableBw {
			availableBw = availableDiskBw
		}

		if float64(req.BandwidthMB) <= bandwidthReserved*availableBw {
			return nodeCount
		}
	}

	return maxNodes
}
