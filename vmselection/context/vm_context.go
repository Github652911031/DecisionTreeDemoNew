package context

import (
	"DecisionTreeDemoNew/vmselection/domain"
	"time"
)

// FilterTraceEntry 过滤执行轨迹
type FilterTraceEntry struct {
	FilterName  string    `json:"filter_name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	InputCount  int       `json:"input_count"`
	OutputCount int       `json:"output_count"`
}

// NodeTraceEntry 节点执行轨迹
type NodeTraceEntry struct {
	NodeKey     string    `json:"node_key"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ResultCount int       `json:"result_count"`
	DurationMs  int64     `json:"duration_ms"`
}

// VMContext 统一执行上下文
type VMContext struct {
	// 过滤链流控
	proceed bool
	jump    bool

	// ===== 输入参数 =====
	Architecture string
	BandwidthMB  int
	VipCount     int
	DiskCount    int
	DeployMode   domain.DeployMode

	// ===== 过滤阶段状态 =====
	CandidateSpecs []domain.VMSpec

	// ===== 计算中间结果 =====
	WriteIOPS     int // 写IOPS总需求
	ReadIOPS      int // 读IOPS总需求
	SetattrDemand int // setattr总需求

	// ===== 执行轨迹 =====
	FilterTrace []FilterTraceEntry
	NodeTrace   []NodeTraceEntry
}

// NewVMContext 创建新的上下文
func NewVMContext(input domain.SelectionInput) *VMContext {
	return &VMContext{
		proceed:      true,
		Architecture: input.Architecture,
		BandwidthMB:  input.BandwidthMB,
		VipCount:     input.VipCount,
		DiskCount:    input.DiskCount,
		DeployMode:   input.DeployMode,
		FilterTrace:  make([]FilterTraceEntry, 0),
		NodeTrace:    make([]NodeTraceEntry, 0),
	}
}

// ===== 实现 ContextWithFlow 接口 =====

// SetProceed 设置是否继续执行
func (c *VMContext) SetProceed(v bool) {
	c.proceed = v
}

// Proceed 获取是否继续执行
func (c *VMContext) Proceed() bool {
	return c.proceed
}

// SetJump 设置是否跳过当前处理器
func (c *VMContext) SetJump(v bool) {
	c.jump = v
}

// Jump 获取是否跳过当前处理器
func (c *VMContext) Jump() bool {
	return c.jump
}

// ===== 轨迹记录方法 =====

// RecordFilterStart 记录过滤器开始执行
func (c *VMContext) RecordFilterStart(filterName string) {
	c.FilterTrace = append(c.FilterTrace, FilterTraceEntry{
		FilterName: filterName,
		StartTime:  time.Now(),
		InputCount: len(c.CandidateSpecs),
	})
}

// RecordFilterEnd 记录过滤器执行完成
func (c *VMContext) RecordFilterEnd(filterName string, outputCount int) {
	if len(c.FilterTrace) == 0 {
		return
	}
	last := &c.FilterTrace[len(c.FilterTrace)-1]
	if last.FilterName == filterName {
		last.EndTime = time.Now()
		last.OutputCount = outputCount
	}
}

// RecordNodeStart 记录节点开始执行
func (c *VMContext) RecordNodeStart(nodeKey string) {
	c.NodeTrace = append(c.NodeTrace, NodeTraceEntry{
		NodeKey:   nodeKey,
		StartTime: time.Now(),
	})
}

// RecordNodeEnd 记录节点执行完成
func (c *VMContext) RecordNodeEnd(nodeKey string, resultCount int) {
	if len(c.NodeTrace) == 0 {
		return
	}
	last := &c.NodeTrace[len(c.NodeTrace)-1]
	if last.NodeKey == nodeKey {
		last.EndTime = time.Now()
		last.ResultCount = resultCount
		last.DurationMs = last.EndTime.Sub(last.StartTime).Milliseconds()
	}
}
