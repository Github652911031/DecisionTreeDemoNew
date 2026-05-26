package domain

import "errors"

var (
	// ErrNoSpecFound 没有找到满足条件的虚机规格
	ErrNoSpecFound = errors.New("没有找到满足条件的虚机规格")

	// ErrInvalidDeployMode 不支持的部署模式
	ErrInvalidDeployMode = errors.New("不支持的部署模式")

	// ErrInvalidQoSStrategy 不支持的QoS策略
	ErrInvalidQoSStrategy = errors.New("不支持的QoS策略")

	// ErrInvalidComponent 不支持的组件类型
	ErrInvalidComponent = errors.New("不支持的组件类型")

	// ErrConfigNotFound 节点配置未找到
	ErrConfigNotFound = errors.New("节点配置未找到")

	// ErrNodeNotInitialized 节点未初始化
	ErrNodeNotInitialized = errors.New("节点未初始化")
)
