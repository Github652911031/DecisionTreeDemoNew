# VM Selection - 虚机选择算法

基于决策树 + 责任链模式实现的虚机规格选择算法。

## 架构设计

### 核心模式组合

1. **决策树模式**：根据部署模式路由到三进程单独部署或 NAS/SPACE 合部部署计算节点
2. **责任链模式**：每个节点内部通过生命周期方法链式执行
3. **配置驱动**：所有算法参数通过 JSON 配置存储在节点上

### 决策树结构

```
RootNode (部署模式路由)
├── SeparateAllBranch
│   ├── MDSNode           - 元数据服务
│   ├── NASNode           - 前端接入服务
│   └── SPACENode         - 数据存储服务
└── SeparateNSBranch
    ├── MDSNode           - 元数据服务
    └── NASSPACENode      - NAS/SPACE 合部服务
```

## 核心数据结构

### VMSpec - 虚机规格

```go
type VMSpec struct {
    ID           int64
    Name         string  // 规格名称，如 c7.large.2
    CPUCores     int     // CPU核数
    MemoryGiB    int     // 内存GiB
    NetBwMBps    int     // 网络带宽 MBps
    DiskBwMBps   int     // 磁盘带宽 MBps
    Architecture string  // x86_64 / aarch64
    NicCount     int     // 支持网卡数量
}
```

### SelectionInput - 选择输入

```go
type SelectionInput struct {
    Architecture string  // 架构: x86_64 / aarch64
    Bandwidth    int     // 文件系统总带宽 MBps
    VipCount     int     // VIP数量
    DiskCount    int     // 数据磁盘总数
    DeployMode   string  // 部署模式: combined / separate
    QoSStrategy  string  // QoS策略: performance / cost / balance
    Component    string  // 组件类型: mds / nas / space
}
```

### NodeConfig - 节点配置

```go
type NodeConfig struct {
    MinCPU              int              // 起步CPU核数
    MinMemory           int              // 起步内存GiB
    UnevennessLimit     float64          // 不均匀度限制
    BandwidthReserved   float64          // 带宽预留比例
    PerformanceTier     map[string]int   // CPU -> 最大能力映射
}
```

## 两种配置模式

### 1. 内存模式（默认）

直接使用代码中硬编码的配置，适合快速开发、测试场景。

```go
specs := []domain.VMSpec{...}
executor := pipeline.NewVMSelectionExecutor(specs) // 默认使用内存树
```

### 2. 数据库模式（推荐）

从 MySQL 数据库动态加载决策树配置，支持热更新。

```go
// 1. 创建数据库连接
db, err := repository.NewDB(repository.DBConfig{
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: os.Getenv("VM_SELECTION_DB_PASSWORD"),
    DBName:   "vm_selection",
})

// 2. 创建工厂
repo := repository.NewTreeRepository(db)
dbFactory := factory.NewDBTreeFactory(repo)

// 3. 创建执行器
executor := pipeline.NewVMSelectionExecutor(
    specs,
    pipeline.WithDBTree("vm_selection", dbFactory),
)

// 4. 支持热重载
executor.ReloadTreeFromDB()
```

## 数据库表设计

| 表名 | 用途 |
|------|------|
| `decision_tree` | 决策树元数据（树ID、名称、是否激活） |
| `decision_tree_node` | **核心配置表** - 每个节点的JSON配置 |

### decision_tree_node 表结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT | 主键 |
| `tree_id` | BIGINT | 所属决策树ID |
| `node_key` | VARCHAR(32) | 节点唯一标识 |
| `node_name` | VARCHAR(64) | 节点名称 |
| `node_type` | VARCHAR(16) | router / leaf |
| `parent_node_id` | BIGINT | 父节点ID |
| `node_config` | JSON | 节点配置JSON |
| `handler_class` | VARCHAR(128) | 处理类名称 |
| `is_enabled` | BOOLEAN | 是否启用 |

## 使用示例

### 合部模式 - 均衡策略

```go
specs := []domain.VMSpec{...}
executor := pipeline.NewVMSelectionExecutor(specs)

input := domain.SelectionInput{
    Architecture: "x86_64",
    Bandwidth:    500,
    VipCount:     10,
    DiskCount:    12,
    DeployMode:   "combined",
    QoSStrategy:  "balance",
}

results, ctx, err := executor.Execute(input)
for _, r := range results {
    fmt.Printf("规格: %s, 节点数: %d, 总CPU: %d\n",
        r.SpecName, r.NodeCount, r.TotalCPU)
}

// 查看执行轨迹
for _, trace := range ctx.NodeTrace {
    fmt.Printf("节点: %s, 结果数: %d\n", trace.NodeKey, trace.ResultCount)
}
```

### 单独模式 - MDS 组件

```go
input := domain.SelectionInput{
    Architecture: "x86_64",
    Bandwidth:    500,
    VipCount:     10,
    DiskCount:    12,
    DeployMode:   "separate",
    Component:    "mds",
}

results, _, err := executor.Execute(input)
```

## 部署模式说明

### Combined 合部模式

NAS 和 SPACE 组件部署在同一批虚机上，MDS 单独部署。

适用场景：
- 中小规模集群
- 资源利用率优先
- 运维简化

### Separate 单独模式

所有组件（MDS、NAS、SPACE）都部署在独立的虚机上。

适用场景：
- 大规模集群
- 性能隔离要求高
- 弹性伸缩需求

## QoS策略说明

### Performance 性能优先

- 更严格的不均匀度限制 (0.25)
- 更高的带宽预留比例 (0.9)
- 单节点承载更少磁盘 (8个)
- 适合高性能场景

### Cost 成本优先

- 较宽松的不均匀度限制 (0.35)
- 较低的带宽预留比例 (0.95)
- 单节点承载更多磁盘 (12个)
- 适合成本敏感场景

### Balance 均衡策略

- 适中的不均匀度限制 (0.3)
- 适中的带宽预留比例 (0.9)
- 单节点承载适中磁盘 (10个)
- 默认策略，平衡性能与成本

## 目录结构

```
vmselection/
├── domain/           # 领域模型
│   ├── models.go     # 数据结构定义
│   └── errors.go     # 错误定义
├── context/          # 执行上下文
│   └── vm_context.go
├── filter/           # 过滤器链
│   ├── base_filter.go
│   ├── architecture_filter.go
│   ├── nic_filter.go
│   └── cpu_memory_filter.go
├── tree/             # 决策树节点
│   ├── base_node.go     # 节点基类
│   ├── root_node.go     # 路由节点
│   ├── mds_node.go      # MDS计算节点
│   ├── nas_node.go      # NAS计算节点
│   ├── space_node.go    # SPACE计算节点
│   └── ns_node.go       # NAS/SPACE合部计算节点
├── factory/          # 工厂层
│   ├── tree_factory.go   # 内存树工厂
│   └── db_tree_factory.go # 数据库树工厂
├── repository/       # 数据访问层
│   ├── models.go     # 数据库实体
│   ├── db.go         # 数据库连接
│   └── tree_repository.go
├── pipeline/         # 执行器
│   └── executor.go
├── example_test.go   # 使用示例
└── vm_selection_test.go  # 单元测试
```

## 数据库初始化

执行 SQL 脚本创建表结构：

```bash
mysql -u root -p < docs/dev-ops/mysql/sql/init.sql
```

或者使用代码自动迁移：

```go
import "DecisionTreeDemoNew/vmselection/repository"

db, _ := repository.NewDB(config)
repository.InitTables(db) // 自动创建/迁移表
```
