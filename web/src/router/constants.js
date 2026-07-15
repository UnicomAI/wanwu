export const PERMS = {
  ADMIN_CENTER: 'admin_center', // 管理员中心（用户、组织、角色有此权限则展示）
  SETTING: 'admin_center.setting', // 管理员中心-平台配置
  OAUTH: 'admin_center.oauth', // 管理员中心-OAuth密钥管理

  OPEN_SOURCE: 'open_source', // 开源仓库

  WGA: 'wga', // 通用智能体
  WGA_WANWU_BOT: 'wga.wanwu_bot', // 通用智能体-WanwuBot
  WGA_OPENCLAW: 'wga.openclaw', // 通用智能体-OpenClaw

  ONTOLOGY: 'ontology', // 本体智能体
  ONTOLOGY_KNOWLEDGE_NETWORK: 'ontology.knowledge_network', // 本体智能体-知识网络
  ONTOLOGY_DATA_SOURCE: 'ontology.data_source', // 本体智能体-数据连接

  MODEL_SERVICE: 'model', // 模型服务
  MODEL_MANAGE: 'model.model_management', // 模型服务-模型管理

  RESOURCE: 'resource', // 资源库
  KNOWLEDGE: 'resource.knowledge', // 资源库-知识库
  MCP_SERVICE: 'resource.mcp', // 资源库-MCP服务
  TOOL: 'resource.tool', // 资源库-工具
  PROMPT: 'resource.prompt', // 资源库-提示词
  SKILL: 'resource.skill', // 资源库-Skill
  SAFETY: 'resource.safety', // 资源库-安全护栏

  APP_SPACE: 'app', // 应用开发
  RAG: 'app.rag', // 应用开发-文本问答
  WORKFLOW: 'app.workflow', // 应用开发-工作流
  AGENT: 'app.agent', // 应用开发-智能体

  SQUARE: 'exploration', // 探索广场
  EXPLORE: 'exploration.app', // 探索广场-应用广场
  MCP: 'exploration.mcp', // 探索广场-MCP广场
  TEMPLATE: 'exploration.template', // 探索广场-模板广场
  SKILL_SQUARE: 'exploration.skill', // 探索广场-Skill广场

  OPERATION: 'operation', // 运营管理
  STATISTIC: 'operation.statistic_client', // 运营管理-统计分析

  APP_OBSERVATION: 'app_observability', // 应用观测
  OBSERVATION_STATISTIC: 'app_observability.statistic', // 应用观测-统计看板

  API_KEY: 'api_key', // API Key管理
  API_KEY_MANAGE: 'api_key.api_key_management', // API Key管理-API Key管理
};
