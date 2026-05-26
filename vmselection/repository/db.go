package repository

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DBConfig 数据库配置
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewDB 创建数据库连接
func NewDB(config DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// InitTables 初始化表结构
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&DecisionTreeEntity{},
		&DecisionTreeNodeEntity{},
	)
}
