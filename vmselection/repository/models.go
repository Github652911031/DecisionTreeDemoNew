package repository

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// DecisionTreeEntity 决策树实体
type DecisionTreeEntity struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TreeKey    string    `gorm:"column:tree_key;uniqueIndex;size:32;not null"`
	TreeName   string    `gorm:"column:tree_name;size:64;not null"`
	TreeDesc   string    `gorm:"column:tree_desc;size:256"`
	RootNodeID int64     `gorm:"column:root_node_id"`
	IsActive   bool      `gorm:"column:is_active;default:true"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (DecisionTreeEntity) TableName() string {
	return "decision_tree"
}

// DecisionTreeNodeEntity 决策树节点实体
type DecisionTreeNodeEntity struct {
	ID                  int64          `gorm:"column:id;primaryKey;autoIncrement"`
	TreeID              int64          `gorm:"column:tree_id;not null;uniqueIndex:idx_tree_node"`
	NodeKey             string         `gorm:"column:node_key;size:32;not null;uniqueIndex:idx_tree_node"`
	NodeName            string         `gorm:"column:node_name;size:64;not null"`
	NodeType            string         `gorm:"column:node_type;size:16;not null"`
	ParentNodeID        int64          `gorm:"column:parent_node_id;index"`
	NodeOrder           int            `gorm:"column:node_order;default:0"`
	NodeConfig          JSONRawMessage `gorm:"column:node_config;type:json"`
	RouteConditionType  string         `gorm:"column:route_condition_type;size:16"`
	RouteConditionValue string         `gorm:"column:route_condition_value;size:32"`
	HandlerClass        string         `gorm:"column:handler_class;size:128"`
	IsEnabled           bool           `gorm:"column:is_enabled;default:true"`
	CreatedAt           time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (DecisionTreeNodeEntity) TableName() string {
	return "decision_tree_node"
}

// JSONRawMessage 自定义JSON类型，用于GORM的JSON字段
type JSONRawMessage json.RawMessage

// Value 实现driver.Valuer接口
func (j JSONRawMessage) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan 实现sql.Scanner接口
func (j *JSONRawMessage) Scan(value interface{}) error {
	if value == nil {
		*j = JSONRawMessage(json.RawMessage(nil))
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSON value, unexpected type %T", value)
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	*j = JSONRawMessage(result)
	return nil
}
