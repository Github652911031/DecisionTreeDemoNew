package factory

import (
	"encoding/json"
	"fmt"

	"DecisionTreeDemoNew/vmselection/repository"
	"DecisionTreeDemoNew/vmselection/tree"
)

// DBTreeFactory 基于数据库的决策树工厂
type DBTreeFactory struct {
	repo *repository.TreeRepository
}

// NewDBTreeFactory 创建数据库决策树工厂
func NewDBTreeFactory(repo *repository.TreeRepository) *DBTreeFactory {
	return &DBTreeFactory{repo: repo}
}

// BuildDecisionTree 从数据库构建决策树
func (f *DBTreeFactory) BuildDecisionTree(treeKey string) (*tree.RootNode, error) {
	treeEntity, err := f.repo.GetActiveTreeByKey(treeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	allNodes, err := f.repo.GetAllNodesByTreeID(treeEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	if len(allNodes) == 0 {
		return nil, fmt.Errorf("no nodes found for tree %s", treeKey)
	}

	rootEntity, childrenByParent, err := buildNodeIndex(allNodes)
	if err != nil {
		return nil, err
	}

	rootNode := tree.NewRootNode()
	rootNode.NodeID = rootEntity.ID
	rootNode.NodeKey = rootEntity.NodeKey
	rootNode.NodeName = rootEntity.NodeName
	rootNode.NodeType = rootEntity.NodeType

	for _, child := range childrenByParent[rootEntity.ID] {
		switch child.NodeKey {
		case "separate_all":
			branch, err := f.buildSeparateAllBranch(child, childrenByParent)
			if err != nil {
				return nil, err
			}
			rootNode.SeparateAll = branch
		case "separate_ns":
			branch, err := f.buildSeparateNSBranch(child, childrenByParent)
			if err != nil {
				return nil, err
			}
			rootNode.SeparateNS = branch
		default:
			return nil, fmt.Errorf("unsupported root child node %q", child.NodeKey)
		}
	}

	if rootNode.SeparateAll == nil {
		return nil, fmt.Errorf("missing required branch separate_all")
	}
	if rootNode.SeparateNS == nil {
		return nil, fmt.Errorf("missing required branch separate_ns")
	}

	if rootNode.SeparateNS.MDSNode == nil {
		rootNode.SeparateNS.MDSNode = rootNode.SeparateAll.MDSNode
	}
	if rootNode.SeparateAll.MDSNode == nil {
		rootNode.SeparateAll.MDSNode = rootNode.SeparateNS.MDSNode
	}

	return rootNode, nil
}

func buildNodeIndex(allNodes []*repository.DecisionTreeNodeEntity) (*repository.DecisionTreeNodeEntity, map[int64][]*repository.DecisionTreeNodeEntity, error) {
	childrenByParent := make(map[int64][]*repository.DecisionTreeNodeEntity)
	var rootEntity *repository.DecisionTreeNodeEntity

	for _, node := range allNodes {
		childrenByParent[node.ParentNodeID] = append(childrenByParent[node.ParentNodeID], node)
		if node.ParentNodeID == 0 || node.NodeKey == "root" {
			rootEntity = node
		}
	}

	if rootEntity == nil {
		return nil, nil, fmt.Errorf("root node not found")
	}

	return rootEntity, childrenByParent, nil
}

func (f *DBTreeFactory) buildSeparateAllBranch(entity *repository.DecisionTreeNodeEntity, childrenByParent map[int64][]*repository.DecisionTreeNodeEntity) (*tree.SeparateAllBranch, error) {
	branch := tree.NewSeparateAllBranch()
	branch.NodeID = entity.ID
	branch.NodeKey = entity.NodeKey
	branch.NodeName = entity.NodeName
	branch.NodeType = entity.NodeType

	for _, child := range childrenByParent[entity.ID] {
		configJSON := json.RawMessage(child.NodeConfig)
		switch child.NodeKey {
		case "mds":
			branch.MDSNode = tree.NewMDSNode(configJSON)
		case "nas":
			branch.NASNode = tree.NewNASNode(configJSON)
		case "space":
			branch.SPACENode = tree.NewSPACENode(configJSON)
		case "ns":
			branch.NASNode = tree.NewNASNode(configJSON)
			branch.SPACENode = tree.NewSPACENode(configJSON)
		}
	}

	return branch, nil
}

func (f *DBTreeFactory) buildSeparateNSBranch(entity *repository.DecisionTreeNodeEntity, childrenByParent map[int64][]*repository.DecisionTreeNodeEntity) (*tree.SeparateNSBranch, error) {
	branch := tree.NewSeparateNSBranch()
	branch.NodeID = entity.ID
	branch.NodeKey = entity.NodeKey
	branch.NodeName = entity.NodeName
	branch.NodeType = entity.NodeType

	for _, child := range childrenByParent[entity.ID] {
		configJSON := json.RawMessage(child.NodeConfig)
		switch child.NodeKey {
		case "mds", "mds_ns":
			branch.MDSNode = tree.NewMDSNode(configJSON)
		case "ns":
			branch.NASSPACENode = tree.NewNASSPACENode(configJSON)
		case "nas":
			branch.NASSPACENode = tree.NewNASSPACENode(configJSON)
		case "space":
			if branch.NASSPACENode == nil {
				branch.NASSPACENode = tree.NewNASSPACENode(configJSON)
			}
		}
	}

	return branch, nil
}
