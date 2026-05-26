package repository

import (
	"fmt"

	"gorm.io/gorm"
)

// TreeRepository 决策树仓库
type TreeRepository struct {
	db *gorm.DB
}

// NewTreeRepository 创建决策树仓库
func NewTreeRepository(db *gorm.DB) *TreeRepository {
	return &TreeRepository{db: db}
}

// GetActiveTreeByKey 根据treeKey获取激活的决策树
func (r *TreeRepository) GetActiveTreeByKey(treeKey string) (*DecisionTreeEntity, error) {
	var tree DecisionTreeEntity
	err := r.db.Where("tree_key = ? AND is_active = ?", treeKey, true).First(&tree).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find tree by key %s: %w", treeKey, err)
	}
	return &tree, nil
}

// GetAllNodesByTreeID 获取树的所有节点
func (r *TreeRepository) GetAllNodesByTreeID(treeID int64) ([]*DecisionTreeNodeEntity, error) {
	var nodes []*DecisionTreeNodeEntity
	err := r.db.Where("tree_id = ? AND is_enabled = ?", treeID, true).
		Order("parent_node_id ASC, node_order ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find nodes for tree %d: %w", treeID, err)
	}
	return nodes, nil
}

// GetRootNode 获取根节点
func (r *TreeRepository) GetRootNode(treeID int64) (*DecisionTreeNodeEntity, error) {
	var node DecisionTreeNodeEntity
	err := r.db.Where("tree_id = ? AND parent_node_id = 0 AND is_enabled = ?", treeID, true).
		First(&node).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find root node for tree %d: %w", treeID, err)
	}
	return &node, nil
}

// GetChildNodes 获取子节点
func (r *TreeRepository) GetChildNodes(parentNodeID int64) ([]*DecisionTreeNodeEntity, error) {
	var nodes []*DecisionTreeNodeEntity
	err := r.db.Where("parent_node_id = ? AND is_enabled = ?", parentNodeID, true).
		Order("node_order ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find child nodes for parent %d: %w", parentNodeID, err)
	}
	return nodes, nil
}
