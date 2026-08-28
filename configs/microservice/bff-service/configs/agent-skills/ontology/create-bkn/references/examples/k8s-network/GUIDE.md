# Kubernetes 模块化网络 - Agent 使用指南

> **网络ID**: k8s-modular  
> **版本**: 1.0.0  
> **标签**: Kubernetes, 拓扑架构, 运维

## 网络概览

本知识网络描述 Kubernetes 集群的核心资源拓扑结构，采用模块化组织方式，每个定义独立成文件。

### 核心对象

| 对象 | 文件路径 | 说明 |
|------|----------|------|
| Pod | `object_types/pod.bkn` | 最小部署单元，包含容器规格 |
| Node | `object_types/node.bkn` | 集群工作节点 |
| Service | `object_types/service.bkn` | 服务暴露与负载均衡 |

### 核心关系

| 关系 | 文件路径 | 源 → 目标 | 说明 |
|------|----------|-----------|------|
| pod_belongs_node | `relation_types/pod_belongs_node.bkn` | Pod → Node | Pod 归属节点关系 |
| service_routes_pod | `relation_types/service_routes_pod.bkn` | Service → Pod | 服务路由到 Pod |

### 可用行动

| 行动 | 文件路径 | 绑定对象 | 说明 |
|------|----------|----------|------|
| restart_pod | `action_types/restart_pod.bkn` | Pod | 重启指定 Pod |
| cordon_node | `action_types/cordon_node.bkn` | Node | 隔离节点，禁止调度 |

## 拓扑结构

```
┌─────────────┐     routes      ┌─────────┐
│  Service    │ ───────────────→│   Pod   │
└─────────────┘                 └────┬────┘
                                     │ belongs
                                     ↓
                                ┌─────────┐
                                │  Node   │
                                └─────────┘
```

## 使用建议

### 查询场景

1. **获取所有 Pod 及其所在节点**
   - 读取 `object_types/pod.bkn` 和 `object_types/node.bkn`
   - 通过 `pod_belongs_node` 关系关联

2. **查找 Service 后端 Pod**
   - 读取 `object_types/service.bkn`
   - 通过 `service_routes_pod` 关系查询目标 Pod

### 运维场景

1. **Pod 故障重启**
   - 使用 `restart_pod` 行动
   - 前置检查：确认 Pod 存在且非关键系统 Pod

2. **节点维护**
   - 使用 `cordon_node` 行动隔离节点
   - 驱逐该节点上的 Pod（如有需要）

## 索引表

### 按类型索引

- **对象定义**: `object_types/*.bkn`
- **关系定义**: `relation_types/*.bkn`
- **行动定义**: `action_types/*.bkn`

### 按功能索引

- **资源管理**: Pod, Node, Service
- **网络路由**: service_routes_pod
- **调度归属**: pod_belongs_node
- **运维操作**: restart_pod, cordon_node

## 注意事项

1. 所有定义文件采用模块化组织，便于独立维护
2. 关系定义中明确指定端点类型和数量约束
3. 行动定义包含前置条件检查，确保操作安全
