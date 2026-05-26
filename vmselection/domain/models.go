package domain

import "fmt"

// ============================================
// 核心数据结构（匹配真实业务）
// ============================================

// VMSpec 虚机规格
type VMSpec struct {
	ID           int64
	Name         string // 规格名称，如 c7.large.2
	CPUCores     int    // CPU核数
	MemoryGiB    int    // 内存GiB
	NetBwGbps    int    // 网络带宽 Gbps (注意：需求是 Gbps！)
	DiskBwGbps   int    // 云硬盘带宽 Gbps
	Architecture string // x86_64 / aarch64
	NicCount     int    // 支持网卡数量
}

// DeployMode 部署模式
type DeployMode string

const (
	// DeployModeSeparateAll 三进程单独部署：mds、nas、space都单独部署
	DeployModeSeparateAll DeployMode = "separate_all"
	// DeployModeSeparateNS NAS/SPACE合部 + MDS单独：mds单独，nas和space合部部署
	DeployModeSeparateNS DeployMode = "separate_ns"
)

// SelectionInput 选择输入（完整业务参数）
type SelectionInput struct {
	Architecture string     // CPU架构: x86_64 / aarch64
	BandwidthMB  int        // 文件系统总带宽 MBps
	VipCount     int        // VIP数量
	DiskCount    int        // EVS云盘总数
	DeployMode   DeployMode // 部署模式
}

// VMQuantity 单角色的虚机规格+数量
type VMQuantity struct {
	SpecName string // 规格名称
	Quantity int    // 需要的节点数量
	CPUCores int    // CPU核数（用于成本排序）
}

// SelectionResult 选择结果（多角色组合输出）
type SelectionResult struct {
	// MDS 角色（固定2台）
	MDS VMQuantity
	// NAS 角色（单独部署模式才有）
	NAS *VMQuantity
	// SPACE 角色（单独部署模式才有）
	SPACE *VMQuantity
	// NASSPACE 合部角色（NAS/SPACE合部模式才有）
	NASSPACE *VMQuantity
	// 总成本 = 各角色CPU核数 × 数量之和
	TotalCost int
}

// ============================================
// 节点配置结构（存储在数据库JSON字段）
// ============================================

// NodeConfig 节点配置（完整业务参数）
type NodeConfig struct {
	// ===== 起步约束 =====
	MinCPU    int `json:"min_cpu"`    // 最小CPU核数
	MinMemory int `json:"min_memory"` // 最小内存GiB
	MinNic    int `json:"min_nic"`    // 最小网卡数量（需求是 >=3）

	// ===== IOPS 密度系数 =====
	WriteIopsDensity float64 `json:"write_iops_density"` // 写IOPS带宽密度（默认16）
	ReadIopsDensity  float64 `json:"read_iops_density"`  // 读IOPS带宽密度（默认48）

	// ===== MDS 专用 =====
	SetattrRatio       float64        `json:"setattr_ratio"`        // setattr占比（默认0.3 = 30%）
	MdsPerformanceTier map[string]int `json:"mds_performance_tier"` // MDS CPU档位 → setattr能力
	MdsNodeCount       int            `json:"mds_node_count"`       // MDS固定节点数（默认2）

	// ===== NAS/SPACE 专用 =====
	BandwidthReserved float64 `json:"bandwidth_reserved"` // 带宽预留比例（默认0.9 = 预留10%）
	UnevennessLimit   float64 `json:"unevenness_limit"`   // 不均匀度上限（默认0.3 = 30%）

	// NAS性能档位 CPU → 单机头读IOPS
	NASPerformanceTierX86 map[string]int `json:"nas_performance_tier_x86"`
	NASPerformanceTierARM map[string]int `json:"nas_performance_tier_arm"`

	// SPACE性能档位 CPU → 单机头读IOPS
	SPACEPerformanceTierX86 map[string]int `json:"space_performance_tier_x86"`
	SPACEPerformanceTierARM map[string]int `json:"space_performance_tier_arm"`

	// NAS/SPACE合部性能档位
	NSPerformanceTierX86 map[string]int `json:"ns_performance_tier_x86"`
	NSPerformanceTierARM map[string]int `json:"ns_performance_tier_arm"`

	// ===== 白名单 =====
	WhitelistSpecs []string `json:"whitelist_specs"` // 允许的虚机规格白名单
}

// ============================================
// 计算辅助函数
// ============================================

// CeilInt 向上取整: ceil(a / b)
func CeilInt(a, b int) int {
	if b == 0 {
		return 0
	}
	return (a + b - 1) / b
}

// CalcUnevenness 计算不均匀度
// roundup(count / nodes) / (count / nodes) - 1
func CalcUnevenness(count, nodes int) float64 {
	if nodes == 0 || count == 0 {
		return 0
	}
	avgPerNode := float64(count) / float64(nodes)
	maxPerNode := float64(CeilInt(count, nodes))
	return (maxPerNode / avgPerNode) - 1.0
}

// MBpsToGbps MBps 转 Gbps (1 Gbps = 125 MBps)
func MBpsToGbps(mbps int) float64 {
	return float64(mbps) / 125.0
}

// GbpsToMBps Gbps 转 MBps
func GbpsToMBps(gbps int) int {
	return gbps * 125
}

// 为了兼容性，保留旧的字段名访问方式
func (s VMSpec) NetBwMBps() int {
	return GbpsToMBps(s.NetBwGbps)
}

func (s VMSpec) DiskBwMBps() int {
	return GbpsToMBps(s.DiskBwGbps)
}

// String 方便调试输出
func (r SelectionResult) String() string {
	result := fmt.Sprintf("MDS: (%s, %d)", r.MDS.SpecName, r.MDS.Quantity)
	if r.NAS != nil {
		result += fmt.Sprintf(", NAS: (%s, %d)", r.NAS.SpecName, r.NAS.Quantity)
	}
	if r.SPACE != nil {
		result += fmt.Sprintf(", SPACE: (%s, %d)", r.SPACE.SpecName, r.SPACE.Quantity)
	}
	if r.NASSPACE != nil {
		result += fmt.Sprintf(", NAS/SPACE: (%s, %d)", r.NASSPACE.SpecName, r.NASSPACE.Quantity)
	}
	result += fmt.Sprintf(", 总成本: %dU", r.TotalCost)
	return result
}
