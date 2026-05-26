# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go (1.22) library implementing generic design patterns for building complex business logic systems:
- **Decision Tree Pattern** - Tree-based routing with configurable nodes
- **Chain of Responsibility Pattern** - Sequential processing chain
- **Strategy Pattern** - Dynamic handler selection based on input
- **Filter Chain Pattern** - Pre-processing filters

## Primary Module: vmselection

The main working module is `vmselection/` - a VM selection algorithm based on decision tree + chain of responsibility patterns.

### Architecture Layers

```
vmselection/
├── domain/          # Pure domain models and errors (no dependencies)
│   ├── models.go    # VMSpec, SelectionInput, SelectionResult, NodeConfig
│   └── errors.go    # Domain errors
├── context/         # Execution context with tracing
│   └── vm_context.go
├── filter/          # Filter chain (pre-processing)
│   ├── base_filter.go
│   ├── architecture_filter.go
│   ├── nic_filter.go
│   └── cpu_memory_filter.go
├── tree/            # Decision tree nodes (strategy pattern)
│   ├── base_node.go     # Base node with lifecycle
│   ├── root_node.go     # Deployment mode router
│   ├── mds_node.go      # MDS calculation
│   ├── nas_node.go      # NAS calculation
│   ├── space_node.go    # SPACE calculation
│   └── ns_node.go       # NAS/SPACE combined calculation
├── factory/         # Tree construction factories
│   ├── tree_factory.go   # In-memory hardcoded tree
│   └── db_tree_factory.go # Database-configured tree
├── repository/      # Data access (MySQL via GORM)
│   ├── models.go     # DB entities: DecisionTree, DecisionTreeNode
│   ├── db.go         # DB connection
│   └── tree_repository.go
└── pipeline/        # Entry point
    └── executor.go   # VMSelectionExecutor
```

### Decision Tree Structure

```
RootNode (deploy_mode router)
├── SeparateAllBranch
│   ├── MDSNode           - Metadata service
│   ├── NASNode           - Frontend access
│   └── SPACENode         - Data storage
└── SeparateNSBranch
    ├── MDSNode           - Metadata service
    └── NASSPACENode      - NAS/SPACE combined service
```

### Key Data Structures

**domain.VMSpec** - VM specification: CPU, memory, network, disk, architecture, nic count
**domain.SelectionInput** - Selection criteria: architecture, bandwidth, vips, disks, deploy_mode
**domain.SelectionResult** - Output: spec, node count, total resources
**context.VMContext** - Execution context with tracing (NodeTrace), filtered specs, errors

## Generic Pattern Modules

### tree/ - Generic Strategy Pattern

Core interfaces for strategy pattern implementation:
```go
type StrategyHandler[T, D, R] interface { Apply(T, D) (R, error) }
type StrategyMapper[T, D, R] interface { Get(T, D) (StrategyHandler, error) }
```

**RouterBase** - Routes requests to appropriate handlers based on StrategyMapper.

### link/model1/ - Chain of Responsibility Pattern

Generic chain pattern:
```go
type LogicLink[T, D, R] interface {
    Apply(T, D) (R, error)
    AppendNext(LogicLink) LogicLink
    Next() LogicLink
}
```

Use BaseLogicLink as base struct for chain nodes, call CallNext() to continue.

### link/model2/ - Business Chain Pattern

Variant with business-specific context and response types.

## Common Commands

```bash
# go path
C:\Users\86188\go\pkg\mod\golang.org\toolchain@v0.0.1-go1.25.8.windows-amd64

# Run all tests
go test ./...

# Run specific test
go test ./vmselection/ -run TestCombinedBalanceMode -v
go test ./vmselection/ -run TestSeparateMDSMode -v
go test ./vmselection/ -run TestDBTreeFactory -v

# Run tests with coverage
go test ./vmselection/ -cover

# Build
go build ./...

# Format code
go fmt ./...
```

## Development Workflow

### Adding a New VM Selection Node

1. Create new file in `vmselection/tree/` e.g. `new_strategy_node.go`
2. Embed `BaseNode` struct
3. Implement `Apply(domain.SelectionInput, *context.VMContext)` method
4. Register in `factory/tree_factory.go` (for in-memory) and `factory/db_tree_factory.go` (for DB mode)
5. Add test case in `vmselection/vm_selection_test.go`

### Adding a New Filter

1. Create in `vmselection/filter/` 
2. Implement `Filter()` method returning `*FilterResult`
3. Register in `BaseNode.Apply()` filter chain

### Database Mode

The `decision_tree_node.node_config` JSON field stores all algorithm parameters. Use SQLite in tests:

```go
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
repository.InitTables(db)
```

SQL initialization script: `docs/dev-ops/mysql/sql/init.sql`

## Reference Files

- Example usage: `vmselection/example_test.go`
- Full test suite: `vmselection/vm_selection_test.go`
- Tree examples: `examples/treeexample/`
- Chain examples: `examples/linkmodel1/`, `examples/linkmodel2/`

## Key Files to Read First

When starting work:
1. `vmselection/README.md` - Complete module documentation
2. `vmselection/domain/models.go` - Understand data structures
3. `vmselection/pipeline/executor.go` - Main entry point
4. `vmselection/tree/base_node.go` - Node lifecycle
5. `tree/strategy.go`, `tree/router.go` - Generic pattern foundation
