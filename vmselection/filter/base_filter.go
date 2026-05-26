package filter

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
)

// ============================================
// 过滤链接口定义
// ============================================

// Filter 过滤器接口
type Filter interface {
	// Apply 执行过滤，返回过滤后的规格列表
	Apply(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error)
}

// BaseFilter 基础过滤器实现
type BaseFilter struct {
	FilterName string
	IsEnabled  bool
}

// ============================================
// 4个核心过滤器（按执行顺序）
// ============================================

// WhitelistFilter 白名单过滤器
type WhitelistFilter struct {
	BaseFilter
	whitelist map[string]bool
}

func NewWhitelistFilter(whitelist []string) *WhitelistFilter {
	wlMap := make(map[string]bool)
	for _, s := range whitelist {
		wlMap[s] = true
	}
	return &WhitelistFilter{
		BaseFilter: BaseFilter{
			FilterName: "WhitelistFilter",
			IsEnabled:  len(whitelist) > 0,
		},
		whitelist: wlMap,
	}
}

func (f *WhitelistFilter) Apply(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	if !f.IsEnabled {
		return specs, nil
	}

	ctx.RecordFilterStart(f.FilterName)
	var filtered []domain.VMSpec
	for _, spec := range specs {
		if f.whitelist[spec.Name] {
			filtered = append(filtered, spec)
		}
	}
	ctx.RecordFilterEnd(f.FilterName, len(filtered))
	return filtered, nil
}

// ArchitectureFilter 架构过滤器
type ArchitectureFilter struct {
	BaseFilter
}

func NewArchitectureFilter() *ArchitectureFilter {
	return &ArchitectureFilter{
		BaseFilter: BaseFilter{
			FilterName: "ArchitectureFilter",
			IsEnabled:  true,
		},
	}
}

func (f *ArchitectureFilter) Apply(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	ctx.RecordFilterStart(f.FilterName)
	var filtered []domain.VMSpec
	for _, spec := range specs {
		if spec.Architecture == ctx.Architecture {
			filtered = append(filtered, spec)
		}
	}
	ctx.RecordFilterEnd(f.FilterName, len(filtered))
	return filtered, nil
}

// NicFilter 网卡数量过滤器
type NicFilter struct {
	BaseFilter
	minNic int
}

func NewNicFilter(minNic int) *NicFilter {
	return &NicFilter{
		BaseFilter: BaseFilter{
			FilterName: "NicFilter",
			IsEnabled:  true,
		},
		minNic: minNic,
	}
}

func (f *NicFilter) Apply(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	ctx.RecordFilterStart(f.FilterName)
	var filtered []domain.VMSpec
	for _, spec := range specs {
		if spec.NicCount >= f.minNic {
			filtered = append(filtered, spec)
		}
	}
	ctx.RecordFilterEnd(f.FilterName, len(filtered))
	return filtered, nil
}

// CPUMemoryFilter CPU内存起步过滤器
type CPUMemoryFilter struct {
	BaseFilter
	minCPU    int
	minMemory int
}

func NewCPUMemoryFilter(minCPU, minMemory int) *CPUMemoryFilter {
	return &CPUMemoryFilter{
		BaseFilter: BaseFilter{
			FilterName: "CPUMemoryFilter",
			IsEnabled:  true,
		},
		minCPU:    minCPU,
		minMemory: minMemory,
	}
}

func (f *CPUMemoryFilter) Apply(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	ctx.RecordFilterStart(f.FilterName)
	var filtered []domain.VMSpec
	for _, spec := range specs {
		if spec.CPUCores >= f.minCPU && spec.MemoryGiB >= f.minMemory {
			filtered = append(filtered, spec)
		}
	}
	ctx.RecordFilterEnd(f.FilterName, len(filtered))
	return filtered, nil
}

// ============================================
// 过滤链执行器
// ============================================

// FilterChain 过滤器链
type FilterChain struct {
	filters []Filter
}

func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]Filter, 0),
	}
}

func (fc *FilterChain) AddFilter(filter Filter) {
	fc.filters = append(fc.filters, filter)
}

func (fc *FilterChain) ApplyAll(specs []domain.VMSpec, ctx *context.VMContext) ([]domain.VMSpec, error) {
	var err error
	result := specs
	for _, filter := range fc.filters {
		result, err = filter.Apply(result, ctx)
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			return nil, domain.ErrNoSpecFound
		}
	}
	return result, nil
}
