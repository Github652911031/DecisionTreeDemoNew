/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

-- 创建数据库（如果不存在）
CREATE database if NOT EXISTS `vm_selection` default character set utf8mb4;
USE vm_selection;

-- ============================================
-- 1. 虚机规格表
-- ============================================
DROP TABLE IF EXISTS vm_flavor_spec;
CREATE TABLE vm_flavor_spec (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    flavor_name VARCHAR(64) NOT NULL COMMENT '规格名称，如 c7.large.2',
    cpu_cores INT NOT NULL COMMENT 'CPU核数',
    memory_gib INT NOT NULL COMMENT '内存GiB',
    net_bw_mbps INT COMMENT '网络带宽 MBps',
    disk_bw_mbps INT COMMENT '磁盘带宽 MBps',
    architecture VARCHAR(16) NOT NULL COMMENT 'CPU架构: x86_64 / aarch64',
    nic_count INT DEFAULT 1 COMMENT '支持网卡数量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_flavor (flavor_name),
    INDEX idx_arch (architecture)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='虚机规格表';

-- 初始化示例规格数据
INSERT INTO vm_flavor_spec (flavor_name, cpu_cores, memory_gib, net_bw_mbps, disk_bw_mbps, architecture, nic_count) VALUES
('c7.large.2', 4, 8, 100, 200, 'x86_64', 2),
('c7.xlarge.4', 8, 16, 200, 400, 'x86_64', 4),
('c7.2xlarge.8', 16, 32, 400, 800, 'x86_64', 8),
('c7.4xlarge.16', 32, 64, 800, 1600, 'x86_64', 16),
('g7.large.2', 4, 8, 100, 200, 'aarch64', 2),
('g7.xlarge.4', 8, 16, 200, 400, 'aarch64', 4),
('g7.2xlarge.8', 16, 32, 400, 800, 'aarch64', 8);

-- ============================================
-- 2. 决策树定义表
-- ============================================
DROP TABLE IF EXISTS decision_tree;
CREATE TABLE decision_tree (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    tree_key VARCHAR(32) NOT NULL COMMENT '树唯一标识: vm_selection',
    tree_name VARCHAR(64) NOT NULL COMMENT '树名称',
    tree_desc VARCHAR(256) COMMENT '树描述',
    root_node_id BIGINT COMMENT '根节点ID',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_tree_key (tree_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='决策树定义表';

-- 初始化决策树
INSERT INTO decision_tree (tree_key, tree_name, tree_desc, root_node_id) VALUES
('vm_selection', '虚机选择决策树', '根据部署模式选择合适的虚机规格', 1);

-- ============================================
-- 3. 决策树节点表（核心配置表）
-- ============================================
DROP TABLE IF EXISTS decision_tree_node;
CREATE TABLE decision_tree_node (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    tree_id BIGINT NOT NULL COMMENT '所属树ID',
    node_key VARCHAR(32) NOT NULL COMMENT '节点唯一标识',
    node_name VARCHAR(64) NOT NULL COMMENT '节点名称',
    node_type VARCHAR(16) NOT NULL COMMENT '节点类型: router(路由)/leaf(计算叶子)',
    parent_node_id BIGINT COMMENT '父节点ID',
    node_order INT DEFAULT 0 COMMENT '同层级节点顺序',

    -- 节点配置JSON，包含起步约束、算法参数、性能档位
    node_config JSON COMMENT '节点配置 {min_cpu, min_memory, bandwidth_reserved, performance_tier:{...}}',

    -- 路由条件（仅router节点需要）
    route_condition_type VARCHAR(16) COMMENT '路由条件类型: deploy_mode/qos_strategy/component',
    route_condition_value VARCHAR(32) COMMENT '触发此节点的条件值',

    handler_class VARCHAR(128) COMMENT '处理类名称，用于工厂反射创建',
    is_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_node (tree_id, node_key),
    INDEX idx_parent (parent_node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='决策树节点表';

-- 初始化决策树节点数据
-- L1: 根节点 - 部署模式路由
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(1, 1, 'root', '根节点', 'router', NULL, 0, NULL, 'deploy_mode', 'RootNode', 1);

-- L2: 部署模式节点
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(2, 1, 'separate_all', '三进程单独部署分支', 'router', 1, 1, NULL, NULL, 'SeparateAllBranch', 1),
(3, 1, 'separate_ns', 'NAS/SPACE合部分支', 'router', 1, 2, NULL, NULL, 'SeparateNSBranch', 1);

-- L3: 三进程单独部署 - MDS
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(4, 1, 'mds', 'MDS元数据服务', 'leaf', 2, 1,
'{
    \"min_cpu\": 4,
    \"min_memory\": 16,
    \"write_iops_density\": 16,
    \"read_iops_density\": 48,
    \"setattr_ratio\": 0.3,
    \"unevenness_limit\": 0.3,
    \"bandwidth_reserved\": 0.9,
    \"performance_tier\": {\"4\": 10000, \"8\": 20000, \"16\": 40000, \"32\": 80000}
}',
NULL, 'MDSNode', 1);

-- L3: 三进程单独部署 - NAS
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(5, 1, 'nas', 'NAS前端接入服务', 'leaf', 2, 2,
'{
    \"min_cpu\": 4,
    \"min_memory\": 8,
    \"write_iops_density\": 16,
    \"read_iops_density\": 48,
    \"setattr_ratio\": 0.3,
    \"unevenness_limit\": 0.3,
    \"bandwidth_reserved\": 0.9,
    \"performance_tier\": {\"4\": 10000, \"8\": 20000, \"16\": 40000}
}',
NULL, 'NASNode', 1);

-- L3: 三进程单独部署 - SPACE
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(6, 1, 'space', 'SPACE数据存储服务', 'leaf', 2, 3,
'{
    \"min_cpu\": 4,
    \"min_memory\": 8,
    \"write_iops_density\": 16,
    \"read_iops_density\": 48,
    \"setattr_ratio\": 0.3,
    \"unevenness_limit\": 0.3,
    \"bandwidth_reserved\": 0.9,
    \"performance_tier\": {\"4\": 10000, \"8\": 20000, \"16\": 40000}
}',
NULL, 'SPACENode', 1);

-- L3: NAS/SPACE合部部署 - MDS
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(7, 1, 'mds_ns', 'MDS元数据服务', 'leaf', 3, 1,
'{
    \"min_cpu\": 4,
    \"min_memory\": 16,
    \"write_iops_density\": 16,
    \"read_iops_density\": 48,
    \"setattr_ratio\": 0.3,
    \"unevenness_limit\": 0.3,
    \"bandwidth_reserved\": 0.9,
    \"performance_tier\": {\"4\": 10000, \"8\": 20000, \"16\": 40000, \"32\": 80000}
}',
NULL, 'MDSNode', 1);

-- L3: NAS/SPACE合部部署 - NAS/SPACE
INSERT INTO decision_tree_node (id, tree_id, node_key, node_name, node_type, parent_node_id, node_order, node_config, route_condition_type, handler_class, is_enabled) VALUES
(8, 1, 'ns', 'NAS/SPACE合部服务', 'leaf', 3, 2,
'{
    \"min_cpu\": 4,
    \"min_memory\": 8,
    \"write_iops_density\": 16,
    \"read_iops_density\": 48,
    \"setattr_ratio\": 0.3,
    \"unevenness_limit\": 0.3,
    \"bandwidth_reserved\": 0.9,
    \"performance_tier\": {\"4\": 10000, \"8\": 20000, \"16\": 40000}
}',
NULL, 'NASSPACENode', 1);

-- ============================================
-- 4. 虚机选择审计日志表
-- ============================================
DROP TABLE IF EXISTS vm_selection_audit_log;
CREATE TABLE vm_selection_audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    execution_id VARCHAR(64) NOT NULL COMMENT '执行唯一ID',
    architecture VARCHAR(16) NOT NULL COMMENT 'CPU架构',
    bandwidth_mbps INT NOT NULL COMMENT '文件系统总带宽(MBps)',
    vip_count INT NOT NULL COMMENT 'VIP数量',
    disk_count INT NOT NULL COMMENT '数据磁盘总数',
    deploy_mode VARCHAR(32) NOT NULL COMMENT '部署模式: separate_all/separate_ns',
    component VARCHAR(32) COMMENT '组件类型: mds/nas/space/ns',

    input_spec_count INT COMMENT '输入规格数',
    filtered_spec_count INT COMMENT '过滤后规格数',
    selected_flavor VARCHAR(64) COMMENT '最终选中规格',
    selected_count INT COMMENT '选中节点数量',

    filter_trace JSON COMMENT '过滤执行轨迹',
    node_trace JSON COMMENT '节点执行轨迹',
    execution_time_ms INT COMMENT '执行耗时(毫秒)',
    error_msg VARCHAR(512) COMMENT '错误信息',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_execution (execution_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='虚机选择审计日志表';

-- ============================================
-- 5. 过滤器配置表
-- ============================================
DROP TABLE IF EXISTS filter_config;
CREATE TABLE filter_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    filter_key VARCHAR(32) NOT NULL COMMENT '过滤器唯一标识',
    filter_name VARCHAR(64) NOT NULL COMMENT '过滤器名称',
    filter_order INT DEFAULT 0 COMMENT '执行顺序',
    is_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    filter_config JSON COMMENT '过滤器配置参数',
    handler_class VARCHAR(128) COMMENT '处理类名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_filter_key (filter_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='过滤器配置表';

-- 初始化过滤器配置
INSERT INTO filter_config (filter_key, filter_name, filter_order, is_enabled, handler_class) VALUES
('architecture', '架构过滤器', 1, 1, 'ArchitectureFilter'),
('nic_count', '网卡数量过滤器', 2, 1, 'NicFilter'),
('cpu_memory', 'CPU内存起步过滤器', 3, 1, 'CPUMemoryFilter');
