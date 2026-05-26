package vmselection

import (
	vmcontext "DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"DecisionTreeDemoNew/vmselection/factory"
	"DecisionTreeDemoNew/vmselection/pipeline"
	"DecisionTreeDemoNew/vmselection/repository"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	sqlitegorm "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	sqlitedrv "modernc.org/sqlite"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
)

var registerOnce sync.Once

func registerSQLiteDriver() {
	registerOnce.Do(func() {
		sql.Register("sqlite-modernc", &sqlitedrv.Driver{})
	})
}

func sampleSpecs() []domain.VMSpec {
	return []domain.VMSpec{
		{
			ID:           1,
			Name:         "c7.large.2",
			CPUCores:     4,
			MemoryGiB:    8,
			NetBwGbps:    1,
			DiskBwGbps:   2,
			Architecture: "x86_64",
			NicCount:     2,
		},
		{
			ID:           2,
			Name:         "c7.xlarge.4",
			CPUCores:     8,
			MemoryGiB:    16,
			NetBwGbps:    2,
			DiskBwGbps:   4,
			Architecture: "x86_64",
			NicCount:     4,
		},
		{
			ID:           3,
			Name:         "c7.2xlarge.8",
			CPUCores:     16,
			MemoryGiB:    32,
			NetBwGbps:    4,
			DiskBwGbps:   8,
			Architecture: "x86_64",
			NicCount:     4,
		},
	}
}

func logResults(t *testing.T, results []domain.SelectionResult) {
	for _, r := range results {
		t.Logf("  %s", r.String())
	}
}

func seedDatabaseTree(t *testing.T, db *gorm.DB, treeKey string) {
	t.Helper()

	treeEntity := &repository.DecisionTreeEntity{
		TreeKey:    treeKey,
		TreeName:   "测试虚机选择决策树",
		TreeDesc:   "用于单元测试的决策树",
		RootNodeID: 1,
		IsActive:   true,
	}
	if err := db.Create(treeEntity).Error; err != nil {
		t.Fatalf("创建决策树失败: %v", err)
	}

	configs := factory.CreateDefaultConfigs()
	marshalConfig := func(key string) repository.JSONRawMessage {
		b, _ := json.Marshal(configs[key])
		return repository.JSONRawMessage(b)
	}

	nodes := []*repository.DecisionTreeNodeEntity{
		{ID: 1, TreeID: treeEntity.ID, NodeKey: "root", NodeName: "根节点", NodeType: "router", ParentNodeID: 0, NodeOrder: 0, RouteConditionType: "deploy_mode", HandlerClass: "RootNode", IsEnabled: true},
		{ID: 2, TreeID: treeEntity.ID, NodeKey: "separate_all", NodeName: "三进程单独部署分支", NodeType: "router", ParentNodeID: 1, NodeOrder: 1, HandlerClass: "SeparateAllBranch", IsEnabled: true},
		{ID: 3, TreeID: treeEntity.ID, NodeKey: "separate_ns", NodeName: "NAS/SPACE合部分支", NodeType: "router", ParentNodeID: 1, NodeOrder: 2, HandlerClass: "SeparateNSBranch", IsEnabled: true},
		{ID: 5, TreeID: treeEntity.ID, NodeKey: "mds", NodeName: "MDS元数据服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 1, NodeConfig: marshalConfig("mds"), HandlerClass: "MDSNode", IsEnabled: true},
		{ID: 6, TreeID: treeEntity.ID, NodeKey: "nas", NodeName: "NAS接入服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 2, NodeConfig: marshalConfig("nas"), HandlerClass: "NASNode", IsEnabled: true},
		{ID: 7, TreeID: treeEntity.ID, NodeKey: "space", NodeName: "SPACE存储服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 3, NodeConfig: marshalConfig("space"), HandlerClass: "SPACENode", IsEnabled: true},
		{ID: 8, TreeID: treeEntity.ID, NodeKey: "mds_ns", NodeName: "MDS元数据服务", NodeType: "leaf", ParentNodeID: 3, NodeOrder: 1, NodeConfig: marshalConfig("mds"), HandlerClass: "MDSNode", IsEnabled: true},
		{ID: 9, TreeID: treeEntity.ID, NodeKey: "ns", NodeName: "NAS/SPACE合部服务", NodeType: "leaf", ParentNodeID: 3, NodeOrder: 2, NodeConfig: marshalConfig("ns"), HandlerClass: "NASSPACENode", IsEnabled: true},
	}

	for _, node := range nodes {
		if err := db.Create(node).Error; err != nil {
			t.Fatalf("创建节点失败: %v", err)
		}
	}
}

func TestSeparateAllMode(t *testing.T) {
	input := domain.SelectionInput{
		Architecture: "x86_64",
		BandwidthMB:  500,
		VipCount:     10,
		DiskCount:    12,
		DeployMode:   domain.DeployModeSeparateAll,
	}

	executor := pipeline.NewVMSelectionExecutor(sampleSpecs())
	results, ctx, err := executor.Execute(input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("没有找到匹配的规格")
	}
	if results[0].NAS == nil || results[0].SPACE == nil {
		t.Fatal("期望返回 MDS/NAS/SPACE 组合结果")
	}

	t.Logf("找到 %d 个匹配的规格", len(results))
	logResults(t, results)
	for _, trace := range ctx.NodeTrace {
		t.Logf("  节点: %s, 结果数: %d, 耗时: %dms", trace.NodeKey, trace.ResultCount, trace.DurationMs)
	}
}

func TestSeparateNSMode(t *testing.T) {
	input := domain.SelectionInput{
		Architecture: "x86_64",
		BandwidthMB:  500,
		VipCount:     10,
		DiskCount:    12,
		DeployMode:   domain.DeployModeSeparateNS,
	}

	executor := pipeline.NewVMSelectionExecutor(sampleSpecs())
	results, ctx, err := executor.Execute(input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("没有找到匹配的规格")
	}
	if results[0].NASSPACE == nil {
		t.Fatal("期望返回 NASSPACE 结果")
	}

	t.Logf("找到 %d 个匹配的规格", len(results))
	logResults(t, results)
	for _, trace := range ctx.NodeTrace {
		t.Logf("  节点: %s, 结果数: %d, 耗时: %dms", trace.NodeKey, trace.ResultCount, trace.DurationMs)
	}
}

func TestFactory(t *testing.T) {
	tree := factory.BuildDecisionTree()
	if tree == nil {
		t.Fatal("构建决策树失败")
	}
	if tree.SeparateAll == nil {
		t.Fatal("separate_all 分支未创建")
	}
	if tree.SeparateNS == nil {
		t.Fatal("separate_ns 分支未创建")
	}
}

func TestDatabaseTreeMode(t *testing.T) {
	registerSQLiteDriver()
	sqlDB, _ := sql.Open("sqlite-modernc", ":memory:")
	db, err := gorm.Open(sqlitegorm.Dialector{Conn: sqlDB}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&repository.DecisionTreeEntity{}, &repository.DecisionTreeNodeEntity{}); err != nil {
		t.Fatalf("迁移表结构失败: %v", err)
	}

	seedDatabaseTree(t, db, "vm_selection_test")

	repo := repository.NewTreeRepository(db)
	dbFactory := factory.NewDBTreeFactory(repo)
	executor := pipeline.NewVMSelectionExecutor(
		sampleSpecs(),
		pipeline.WithDBTree("vm_selection_test", dbFactory),
	)

	input := domain.SelectionInput{
		Architecture: "x86_64",
		BandwidthMB:  500,
		VipCount:     10,
		DiskCount:    12,
		DeployMode:   domain.DeployModeSeparateAll,
	}

	results, ctx, err := executor.Execute(input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("没有找到匹配的规格")
	}
	if len(ctx.NodeTrace) == 0 {
		t.Fatal("节点执行轨迹为空")
	}

	t.Log("【数据库模式测试】虚机选择成功")
	logResults(t, results)

	results2, _, err := executor.Execute(input)
	if err != nil {
		t.Fatalf("再次执行失败: %v", err)
	}
	if len(results2) != len(results) {
		t.Fatalf("重复执行结果不一致，期望 %d 个，实际 %d 个", len(results), len(results2))
	}

	if err := executor.ReloadTreeFromDB(); err != nil {
		t.Fatalf("热重载失败: %v", err)
	}
}

func TestDatabaseTreeConfigValidation(t *testing.T) {
	registerSQLiteDriver()
	sqlDB, _ := sql.Open("sqlite-modernc", ":memory:")
	db, err := gorm.Open(sqlitegorm.Dialector{Conn: sqlDB}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&repository.DecisionTreeEntity{}, &repository.DecisionTreeNodeEntity{}); err != nil {
		t.Fatalf("迁移表结构失败: %v", err)
	}

	treeEntity := &repository.DecisionTreeEntity{TreeKey: "vm_config_test", TreeName: "配置验证测试树", RootNodeID: 1, IsActive: true}
	if err := db.Create(treeEntity).Error; err != nil {
		t.Fatalf("创建测试树失败: %v", err)
	}

	testCases := []struct {
		name        string
		minCPU      int
		specCPU     int
		shouldMatch bool
	}{
		{"起步4核 - 4核规格匹配", 4, 4, true},
		{"起步8核 - 4核规格不匹配", 8, 4, false},
		{"起步2核 - 4核规格匹配", 2, 4, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := db.Exec("DELETE FROM decision_tree_node").Error; err != nil {
				t.Fatalf("清理节点失败: %v", err)
			}

			mdsConfig := map[string]interface{}{
				"min_cpu":              tc.minCPU,
				"min_memory":           8,
				"min_nic":              2,
				"write_iops_density":   16,
				"setattr_ratio":        0.3,
				"mds_node_count":       2,
				"mds_performance_tier": map[string]int{fmt.Sprintf("%d", tc.minCPU): 10000},
			}
			mdsConfigJSON, _ := json.Marshal(mdsConfig)

			nodes := []*repository.DecisionTreeNodeEntity{
				{ID: 1, TreeID: treeEntity.ID, NodeKey: "root", NodeName: "根节点", NodeType: "router", ParentNodeID: 0, NodeOrder: 0, HandlerClass: "RootNode", IsEnabled: true},
				{ID: 2, TreeID: treeEntity.ID, NodeKey: "separate_all", NodeName: "三进程单独部署分支", NodeType: "router", ParentNodeID: 1, NodeOrder: 1, HandlerClass: "SeparateAllBranch", IsEnabled: true},
				{ID: 3, TreeID: treeEntity.ID, NodeKey: "separate_ns", NodeName: "NAS/SPACE合部分支", NodeType: "router", ParentNodeID: 1, NodeOrder: 2, HandlerClass: "SeparateNSBranch", IsEnabled: true},
				{ID: 4, TreeID: treeEntity.ID, NodeKey: "mds", NodeName: "MDS元数据服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 1, NodeConfig: repository.JSONRawMessage(mdsConfigJSON), HandlerClass: "MDSNode", IsEnabled: true},
				{ID: 5, TreeID: treeEntity.ID, NodeKey: "nas", NodeName: "NAS接入服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 2, NodeConfig: repository.JSONRawMessage(mdsConfigJSON), HandlerClass: "NASNode", IsEnabled: true},
				{ID: 6, TreeID: treeEntity.ID, NodeKey: "space", NodeName: "SPACE存储服务", NodeType: "leaf", ParentNodeID: 2, NodeOrder: 3, NodeConfig: repository.JSONRawMessage(mdsConfigJSON), HandlerClass: "SPACENode", IsEnabled: true},
				{ID: 7, TreeID: treeEntity.ID, NodeKey: "mds_ns", NodeName: "MDS元数据服务", NodeType: "leaf", ParentNodeID: 3, NodeOrder: 1, NodeConfig: repository.JSONRawMessage(mdsConfigJSON), HandlerClass: "MDSNode", IsEnabled: true},
				{ID: 8, TreeID: treeEntity.ID, NodeKey: "ns", NodeName: "NAS/SPACE合部服务", NodeType: "leaf", ParentNodeID: 3, NodeOrder: 2, NodeConfig: repository.JSONRawMessage(mdsConfigJSON), HandlerClass: "NASSPACENode", IsEnabled: true},
			}
			for _, node := range nodes {
				if err := db.Create(node).Error; err != nil {
					t.Fatalf("创建节点失败: %v", err)
				}
			}

			specs := []domain.VMSpec{{ID: 1, Name: "test.spec", CPUCores: tc.specCPU, MemoryGiB: 8, NetBwGbps: 1, DiskBwGbps: 2, Architecture: "x86_64", NicCount: 2}}
			repo := repository.NewTreeRepository(db)
			dbFactory := factory.NewDBTreeFactory(repo)
			executor := pipeline.NewVMSelectionExecutor(specs, pipeline.WithDBTree("vm_config_test", dbFactory))

			input := domain.SelectionInput{Architecture: "x86_64", BandwidthMB: 500, VipCount: 10, DiskCount: 12, DeployMode: domain.DeployModeSeparateAll}
			results, _, err := executor.Execute(input)

			hasMatch := len(results) > 0 && err == nil
			if hasMatch != tc.shouldMatch {
				t.Errorf("期望匹配=%v，实际=%v (起步CPU=%d, 规格CPU=%d, err=%v)", tc.shouldMatch, hasMatch, tc.minCPU, tc.specCPU, err)
			}
		})
	}
}

func TestEndToEndRealMySQL(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("跳过测试: 未设置 MYSQL_DSN")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		t.Skipf("跳过测试: 无法连接 MySQL 数据库 (%v)", err)
		return
	}

	var flavorSpecs []struct {
		FlavorName   string `gorm:"column:flavor_name"`
		CPUCores     int    `gorm:"column:cpu_cores"`
		MemoryGiB    int    `gorm:"column:memory_gib"`
		NetBwGbps    int    `gorm:"column:net_bw_gbps"`
		DiskBwGbps   int    `gorm:"column:disk_bw_gbps"`
		Architecture string `gorm:"column:architecture"`
		NicCount     int    `gorm:"column:nic_count"`
	}
	if err := db.Table("vm_flavor_spec").Where("architecture = ?", "x86_64").Find(&flavorSpecs).Error; err != nil {
		t.Fatalf("读取规格列表失败: %v", err)
	}
	if len(flavorSpecs) == 0 {
		t.Fatal("数据库中没有规格数据，请先执行 init.sql 初始化")
	}

	var specs []domain.VMSpec
	for i, f := range flavorSpecs {
		specs = append(specs, domain.VMSpec{
			ID:           int64(i + 1),
			Name:         f.FlavorName,
			CPUCores:     f.CPUCores,
			MemoryGiB:    f.MemoryGiB,
			NetBwGbps:    f.NetBwGbps,
			DiskBwGbps:   f.DiskBwGbps,
			Architecture: f.Architecture,
			NicCount:     f.NicCount,
		})
	}

	repo := repository.NewTreeRepository(db)
	dbFactory := factory.NewDBTreeFactory(repo)
	executor := pipeline.NewVMSelectionExecutor(specs, pipeline.WithDBTree("vm_selection", dbFactory))

	testScenarios := []struct {
		name       string
		input      domain.SelectionInput
		expectMin  int
		expectNode string
	}{
		{
			name:       "三进程单独部署",
			input:      domain.SelectionInput{Architecture: "x86_64", BandwidthMB: 500, VipCount: 10, DiskCount: 12, DeployMode: domain.DeployModeSeparateAll},
			expectMin:  1,
			expectNode: "space",
		},
		{
			name:       "NAS/SPACE 合部部署",
			input:      domain.SelectionInput{Architecture: "x86_64", BandwidthMB: 800, VipCount: 15, DiskCount: 18, DeployMode: domain.DeployModeSeparateNS},
			expectMin:  1,
			expectNode: "ns",
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			startTime := time.Now()
			results, ctx, err := executor.Execute(scenario.input)
			duration := time.Since(startTime)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}
			if len(results) < scenario.expectMin {
				t.Fatalf("结果数量不足，期望至少 %d，实际 %d", scenario.expectMin, len(results))
			}

			foundNode := false
			for _, trace := range ctx.NodeTrace {
				if trace.NodeKey == scenario.expectNode {
					foundNode = true
					break
				}
			}
			if !foundNode {
				t.Errorf("执行轨迹中未找到期望节点 %s，实际路径: %v", scenario.expectNode, getNodeKeys(ctx.NodeTrace))
			}

			filterTraceJSON, _ := json.Marshal(ctx.FilterTrace)
			nodeTraceJSON, _ := json.Marshal(ctx.NodeTrace)
			auditLog := map[string]interface{}{
				"execution_id":        fmt.Sprintf("test_%d", time.Now().UnixNano()),
				"architecture":        scenario.input.Architecture,
				"bandwidth_mbps":      scenario.input.BandwidthMB,
				"vip_count":           scenario.input.VipCount,
				"disk_count":          scenario.input.DiskCount,
				"deploy_mode":         scenario.input.DeployMode,
				"input_spec_count":    len(specs),
				"filtered_spec_count": len(ctx.CandidateSpecs),
				"selected_result":     results[0].String(),
				"filter_trace":        filterTraceJSON,
				"node_trace":          nodeTraceJSON,
				"execution_time_ms":   duration.Milliseconds(),
				"created_at":          time.Now(),
			}
			if err := db.Table("vm_selection_audit_log").Create(auditLog).Error; err != nil {
				t.Errorf("写入审计日志失败 (非致命): %v", err)
			}
		})
	}
}

func getNodeKeys(traces []vmcontext.NodeTraceEntry) []string {
	keys := make([]string, len(traces))
	for i, t := range traces {
		keys[i] = t.NodeKey
	}
	return keys
}
