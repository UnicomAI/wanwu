# 企查查 MCP 广场卡片

依赖通用 MCP OAuth 功能。广场显示「企查查 MCP · OAuth」卡片；管理员启用
`qcc` provider 后，卡片提供浏览器授权入口，无需客户手动填写 API Key。

## 配置

参照 [通用部署说明](README.md)，复制 `qcc.example.json`，将两个
`https://wanwu.example/...` 地址改为本部署可用的 HTTPS 转发入口、Web 回调地址。
使用 `WANWU_MCP_OAUTH_CONFIG` 加载配置，并通过部署密钥机制设置
`WANWU_MCP_OAUTH_KEY`；不要将密钥或运行时凭据提交到仓库。

`clientId` 留空时，按企查查元数据进行公开客户端动态注册，Client Name 为
**Wanwu**，使用授权码 + PKCE S256。也可以填写为本部署回调预先注册的 client ID。
这不是 API Key，客户无需取得或粘贴它。域名/回调路径必须与注册内容一致。

## 一次授权、五个连接

按企查查 MCP OAuth 接入文档 V1.4（2026-07-30），以下资源属于同一授权集合：

| 连接名称 | Resource |
| --- | --- |
| 企业数据 | `https://agent.qcc.com/mcp/company/stream` |
| 风险数据 | `https://agent.qcc.com/mcp/risk/stream` |
| 知识产权 | `https://agent.qcc.com/mcp/ipr/stream` |
| 经营信息 | `https://agent.qcc.com/mcp/operation/stream` |
| 董监高信息 | `https://agent.qcc.com/mcp/executive/stream` |

使用 company 作为授权及刷新请求的入口 resource，显式允许上述五个目标。
授权完成后保存五条具名 MCP 连接，共用一份授权状态及刷新锁；重复保存跳过已存在
的连接，中断保存可重试。用户可按任务选择需要的连接，不强制把全部工具加载到每次对话。
不根据相同域名推断其他端点可用，也不包含未在文档中列出的资源。

解绑或删除此授权下任一托管连接会停用整组授权。重新授权后点击「保存连接」更新
现有连接的内部引用，旧引用不会恢复有效。访问数据范围由企查查账号实际权限决定；
授权成功不表示所有数据接口都已开通。

## 验收边界

`TestQCCPresetBundle` 校验配置和五个明确的 Resource；
`TestFiveTargetsShareRefreshAndRevocation` 使用隔离的模拟资源服务验证五目标调用、
并发单次刷新、持久化恢复和整体解绑。通用功能另有双 SDK 真实协议请求测试。
这些测试不消耗企查查额度、不包含客户账号，不能代替部署后的短信登录及真实业务查询验收。
