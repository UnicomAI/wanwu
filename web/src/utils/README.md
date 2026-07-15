# 路由 Headers 配置

## 概述

这个功能允许根据不同的路由路径自动设置不同的 HTTP 请求头参数。通过配置路由匹配规则和对应的 headers，系统会自动为不同的功能模块添加相应的标识信息。

## 文件结构

- `routeHeadersConfig.js` - 路由 headers 配置文件
- `request.js` - 请求拦截器，应用路由 headers 配置

## 配置说明

### 配置项结构

```javascript
{
  patterns: ['/agent', '/webChat'],           // 路由匹配模式
  headers: {                                  // 对应的 headers
    'x-module': 'agent',
    'x-feature': 'intelligence',
    'x-ai-enabled': 'true'
  },
  priority: 9,                                // 优先级（数字越大优先级越高）
  description: '智能体模块'                   // 模块描述
}
```

### 优先级规则

- 优先级数字越大，优先级越高
- 高优先级的配置会覆盖低优先级的配置
- 多个匹配的配置会合并，相同字段以高优先级为准

### 当前支持的模块

| 模块         | 路由模式                           | 优先级 | 描述          |
|------------|--------------------------------|-----|-------------|
| preview    | `/doc`, `/txtView`, `/jsExcel` | 10  | 文件预览模块      |
| agent      | `/agent`, `/webChat`           | 9   | 智能体模块       |
| rag        | `/rag`, `/knowledge`           | 8   | RAG 和知识库模块  |
| workflow   | `/workflow`                    | 7   | 工作流模块       |
| mcp        | `/mcp`                         | 6   | MCP 管理模块    |
| permission | `/permission`, `/userCenter`   | 5   | 权限管理和用户中心模块 |
| safety     | `/safety`                      | 4   | 安全护栏模块      |
| model      | `/modelAccess`                 | 3   | 模型访问模块      |
| explore    | `/explore`                     | 2   | 探索广场模块      |
| doc        | `/docCenter`                   | 1   | 文档中心模块      |
| appspace   | `/appSpace`                    | 0   | 应用空间模块      |

## 使用方法

### 1. 基本使用

系统会自动根据当前路由路径设置相应的 headers，无需额外配置。

### 2. 添加新的模块配置

在 `routeHeadersConfig.js` 中添加新的配置项：

```javascript
{
  patterns: ['/newModule'],
  headers: {
    'x-module': 'newmodule',
    'x-feature': 'newfeature',
    'x-custom-header': 'custom-value'
  },
  priority: 5,
  description: '新模块描述'
}
```

### 3. 获取模块 headers

```javascript
import {
  getModuleHeaders,
  getAvailableModules,
} from '@/utils/routeHeadersConfig';

// 获取指定模块的 headers
const agentHeaders = getModuleHeaders('agent');

// 获取所有可用模块列表
const modules = getAvailableModules();
```

## Headers 字段说明

### 通用字段

- `x-module`: 模块标识
- `x-feature`: 功能特性
- `x-cache-control`: 缓存控制策略

### 特殊字段

- `x-ai-enabled`: AI 功能启用状态
- `x-search-enabled`: 搜索功能启用状态
- `x-priority`: 请求优先级
- `x-security`: 安全级别
- `x-validation`: 验证级别
- `x-rate-limit`: 速率限制
- `x-content-type`: 内容类型
- `x-execution-mode`: 执行模式

## 调试信息

在开发环境下，系统会在控制台输出详细的调试信息：

- 路由匹配结果
- 优先级排序
- 最终应用的 headers

## 注意事项

1. **路由匹配**: 使用 `includes()` 进行匹配，支持部分路径匹配
2. **优先级**: 确保设置合理的优先级，避免冲突
3. **性能**: 配置项不宜过多，建议控制在合理范围内
4. **维护**: 定期检查和更新配置，确保与实际路由保持一致

## 扩展功能

### 动态配置

可以通过 API 动态获取配置，支持运行时更新：

```javascript
// 动态更新配置
export const updateRouteHeadersConfig = newConfig => {
  routeHeadersConfig.length = 0;
  routeHeadersConfig.push(...newConfig);
};
```

### 条件配置

支持根据环境变量或其他条件动态调整配置：

```javascript
// 根据环境调整配置
if (process.env.NODE_ENV === 'production') {
  // 生产环境特定配置
}
```

## 故障排除

### 常见问题

1. **Headers 未生效**: 检查路由匹配模式和优先级设置
2. **配置冲突**: 检查是否有重复的路由模式
3. **性能问题**: 检查配置项数量，避免过多匹配规则

### 调试步骤

1. 检查控制台输出的调试信息
2. 验证路由路径是否正确
3. 确认配置项的优先级设置
4. 检查 headers 合并逻辑
