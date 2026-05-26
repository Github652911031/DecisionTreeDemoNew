package tree

import (
	"DecisionTreeDemoNew/tree"
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"DecisionTreeDemoNew/vmselection/filter"
	"encoding/json"
	"sort"
	"strconv"
)

// VMBaseNode 所有虚机选择节点的基类
type VMBaseNode struct {
	NodeID     int64           // 节点ID
	NodeKey    string          // 节点唯一标识
	NodeName   string          // 节点名称
	NodeType   string          // router / leaf
	ConfigJSON json.RawMessage // 节点配置JSON
	Config     domain.NodeConfig
}

// ParseConfig 解析配置JSON
func (n *VMBaseNode) ParseConfig() error {
	if len(n.ConfigJSON) == 0 {
		return domain.ErrConfigNotFound
	}
	return json.Unmarshal(n.ConfigJSON, &n.Config)
}

// ===== 叶子节点基类 =====

// LeafNode 所有计算叶子节点的基类
type LeafNode struct {
	VMBaseNode
}

// ApplyBase 供具体节点调用，执行生命周期流程
func (n *LeafNode) ApplyBase(node tree.MultiThreadNode[domain.SelectionInput, *context.VMContext, []domain.SelectionResult], req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	base := tree.NewMultiThreadRouterBase[domain.SelectionInput, *context.VMContext, []domain.SelectionResult]()
	return base.Apply(node, req, ctx)
}

// ApplyBefore 默认实现：加载配置，记录节点开始
func (n *LeafNode) ApplyBefore(req domain.SelectionInput, ctx *context.VMContext) (tree.ApplyBeforeResult[[]domain.SelectionResult], error) {
	ctx.RecordNodeStart(n.NodeKey)
	if err := n.ParseConfig(); err != nil {
		return tree.ApplyBeforeResult[[]domain.SelectionResult]{}, err
	}
	return tree.ApplyBeforeResult[[]domain.SelectionResult]{}, nil
}

// MultiThread 默认实现：预计算IOPS等中间值
func (n *LeafNode) MultiThread(req domain.SelectionInput, ctx *context.VMContext) error {
	ctx.WriteIOPS = int(float64(req.BandwidthMB) * n.Config.WriteIopsDensity)
	ctx.ReadIOPS = int(float64(req.BandwidthMB) * n.Config.ReadIopsDensity)
	ctx.SetattrDemand = int(float64(ctx.WriteIOPS) * n.Config.SetattrRatio)
	return nil
}

// ApplyAfter 默认实现：记录日志，按总成本排序结果
func (n *LeafNode) ApplyAfter(req domain.SelectionInput, ctx *context.VMContext, result []domain.SelectionResult) error {
	ctx.RecordNodeEnd(n.NodeKey, len(result))
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalCost < result[j].TotalCost
	})
	return nil
}

// ApplyAfterException 默认实现：异常处理
func (n *LeafNode) ApplyAfterException(req domain.SelectionInput, ctx *context.VMContext, err error) error {
	return nil
}

// Get 实现StrategyMapper接口（叶子节点不需要路由，返回nil）
func (n *LeafNode) Get(req domain.SelectionInput, ctx *context.VMContext) (tree.StrategyHandler[domain.SelectionInput, *context.VMContext, []domain.SelectionResult], error) {
	return nil, nil
}

// ApplyFilters 应用过滤链（白名单 → 架构 → 网卡 → CPU内存）
func (n *LeafNode) ApplyFilters(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	fc := filter.NewFilterChain()

	// 1. 白名单过滤
	if len(n.Config.WhitelistSpecs) > 0 {
		fc.AddFilter(filter.NewWhitelistFilter(n.Config.WhitelistSpecs))
	}

	// 2. 架构过滤
	fc.AddFilter(filter.NewArchitectureFilter())

	// 3. 网卡过滤（默认至少3个）
	minNic := 3
	if n.Config.MinNic > 0 {
		minNic = n.Config.MinNic
	}
	fc.AddFilter(filter.NewNicFilter(minNic))

	// 4. CPU内存起步过滤
	fc.AddFilter(filter.NewCPUMemoryFilter(n.Config.MinCPU, n.Config.MinMemory))

	return fc.ApplyAll(specs, ctx)
}

// ===== 路由节点基类 =====

// RouterNode 路由节点基类
type RouterNode struct {
	VMBaseNode
}

// ApplyBase 供具体路由节点调用，执行生命周期流程
func (n *RouterNode) ApplyBase(node tree.MultiThreadNode[domain.SelectionInput, *context.VMContext, []domain.SelectionResult], req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	base := tree.NewMultiThreadRouterBase[domain.SelectionInput, *context.VMContext, []domain.SelectionResult]()
	return base.Apply(node, req, ctx)
}

// ApplyBefore 默认实现
func (n *RouterNode) ApplyBefore(req domain.SelectionInput, ctx *context.VMContext) (tree.ApplyBeforeResult[[]domain.SelectionResult], error) {
	ctx.RecordNodeStart(n.NodeKey)
	return tree.ApplyBeforeResult[[]domain.SelectionResult]{}, nil
}

// MultiThread 路由节点默认不做计算
func (n *RouterNode) MultiThread(req domain.SelectionInput, ctx *context.VMContext) error {
	return nil
}

// DoApply 默认实现：路由到子节点
func (n *RouterNode) DoApply(mapper tree.StrategyMapper[domain.SelectionInput, *context.VMContext, []domain.SelectionResult], req domain.SelectionInput, ctx *context.VMContext) ([]domain.SelectionResult, error) {
	base := tree.NewMultiThreadRouterBase[domain.SelectionInput, *context.VMContext, []domain.SelectionResult]()
	return base.Route(mapper, req, ctx)
}

// ApplyAfter 默认实现
func (n *RouterNode) ApplyAfter(req domain.SelectionInput, ctx *context.VMContext, result []domain.SelectionResult) error {
	ctx.RecordNodeEnd(n.NodeKey, len(result))
	return nil
}

// ApplyAfterException 默认实现
func (n *RouterNode) ApplyAfterException(req domain.SelectionInput, ctx *context.VMContext, err error) error {
	return nil
}

// ===== 通用工具方法 =====

// 过滤满足CPU和内存要求的规格
func FilterSpecsByMinRequirements(specs []domain.VMSpec, minCPU, minMemory int) []domain.VMSpec {
	var filtered []domain.VMSpec
	for _, spec := range specs {
		if spec.CPUCores >= minCPU && spec.MemoryGiB >= minMemory {
			filtered = append(filtered, spec)
		}
	}
	return filtered
}

// 查找满足setattr要求的最小CPU核数
func FindMinCPUForSetattr(performanceTier map[string]int, setattrDemand int) int {
	minCPU := 0
	for cpuStr, maxSetattr := range performanceTier {
		cpu, err := strconv.Atoi(cpuStr)
		if err != nil {
			continue
		}
		if maxSetattr >= setattrDemand && (minCPU == 0 || cpu < minCPU) {
			minCPU = cpu
		}
	}
	return minCPU
}

// 向上取整
func CeilInt(a, b int) int {
	if b == 0 {
		return 0
	}
	return (a + b - 1) / b
}

func SortResultsByCost(results []domain.SelectionResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCost < results[j].TotalCost
	})
}
