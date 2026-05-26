package pipeline

import (
	"DecisionTreeDemoNew/vmselection/context"
	"DecisionTreeDemoNew/vmselection/domain"
	"DecisionTreeDemoNew/vmselection/factory"
	"DecisionTreeDemoNew/vmselection/tree"
)

// VMSelectionExecutor 虚机选择执行器
type VMSelectionExecutor struct {
	specs     []domain.VMSpec
	rootNode  *tree.RootNode
	treeKey   string
	dbFactory *factory.DBTreeFactory
}

// Option 执行器配置选项
type Option func(*VMSelectionExecutor)

// WithInMemoryTree 使用内存中的硬编码决策树
func WithInMemoryTree() Option {
	return func(e *VMSelectionExecutor) {
		e.rootNode = factory.BuildDecisionTree()
	}
}

// WithDBTree 使用数据库加载的决策树
func WithDBTree(treeKey string, dbFactory *factory.DBTreeFactory) Option {
	return func(e *VMSelectionExecutor) {
		e.treeKey = treeKey
		e.dbFactory = dbFactory
	}
}

// NewVMSelectionExecutor 创建执行器
func NewVMSelectionExecutor(specs []domain.VMSpec, opts ...Option) *VMSelectionExecutor {
	e := &VMSelectionExecutor{
		specs: specs,
	}

	// 默认使用内存树
	e.rootNode = factory.BuildDecisionTree()

	// 应用选项
	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Execute 执行虚机选择
func (e *VMSelectionExecutor) Execute(input domain.SelectionInput) ([]domain.SelectionResult, *context.VMContext, error) {
	// 创建上下文
	ctx := context.NewVMContext(input)

	// 设置候选规格
	ctx.CandidateSpecs = e.specs

	// 如果配置了数据库工厂，从数据库加载决策树
	root := e.rootNode
	if e.dbFactory != nil && e.treeKey != "" {
		var err error
		root, err = e.dbFactory.BuildDecisionTree(e.treeKey)
		if err != nil {
			return nil, ctx, err
		}
	}

	// 执行决策树
	result, err := root.Apply(input, ctx)
	if err != nil {
		return nil, ctx, err
	}

	return result, ctx, nil
}

// ReloadTreeFromDB 重新从数据库加载决策树
func (e *VMSelectionExecutor) ReloadTreeFromDB() error {
	if e.dbFactory == nil || e.treeKey == "" {
		return nil
	}

	var err error
	e.rootNode, err = e.dbFactory.BuildDecisionTree(e.treeKey)
	return err
}
