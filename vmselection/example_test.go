package vmselection

import (
	"fmt"
	"log"
	"os"

	"DecisionTreeDemoNew/vmselection/domain"
	"DecisionTreeDemoNew/vmselection/factory"
	"DecisionTreeDemoNew/vmselection/pipeline"
	"DecisionTreeDemoNew/vmselection/repository"
)

// Example_inMemoryMode 内存模式使用示例
func Example_inMemoryMode() {
	specs := []domain.VMSpec{
		{ID: 1, Name: "c7.large.2", CPUCores: 4, MemoryGiB: 8, NetBwGbps: 1, DiskBwGbps: 2, Architecture: "x86_64", NicCount: 2},
		{ID: 2, Name: "c7.xlarge.4", CPUCores: 8, MemoryGiB: 16, NetBwGbps: 2, DiskBwGbps: 4, Architecture: "x86_64", NicCount: 4},
		{ID: 3, Name: "c7.2xlarge.8", CPUCores: 16, MemoryGiB: 32, NetBwGbps: 4, DiskBwGbps: 8, Architecture: "x86_64", NicCount: 4},
	}

	executor := pipeline.NewVMSelectionExecutor(specs)
	input := domain.SelectionInput{
		Architecture: "x86_64",
		BandwidthMB:  500,
		VipCount:     10,
		DiskCount:    12,
		DeployMode:   domain.DeployModeSeparateAll,
	}

	results, ctx, err := executor.Execute(input)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range results {
		fmt.Println(r.String())
	}

	fmt.Println("执行过程:")
	for _, trace := range ctx.NodeTrace {
		fmt.Printf("  节点: %s, 结果数: %d\n", trace.NodeKey, trace.ResultCount)
	}
}

// Example_databaseMode 数据库模式使用示例
func Example_databaseMode() {
	dbConfig := repository.DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: os.Getenv("VM_SELECTION_DB_PASSWORD"),
		DBName:   "vm_selection",
	}

	db, err := repository.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	repo := repository.NewTreeRepository(db)
	dbFactory := factory.NewDBTreeFactory(repo)

	specs := []domain.VMSpec{
		{ID: 1, Name: "c7.xlarge.4", CPUCores: 8, MemoryGiB: 16, NetBwGbps: 2, DiskBwGbps: 4, Architecture: "x86_64", NicCount: 4},
	}

	executor := pipeline.NewVMSelectionExecutor(
		specs,
		pipeline.WithDBTree("vm_selection", dbFactory),
	)

	input := domain.SelectionInput{
		Architecture: "x86_64",
		BandwidthMB:  500,
		VipCount:     10,
		DiskCount:    12,
		DeployMode:   domain.DeployModeSeparateAll,
	}

	results, _, err := executor.Execute(input)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range results {
		fmt.Println(r.String())
	}
}
