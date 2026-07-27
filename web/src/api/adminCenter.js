import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

/**
 * 管理员中心 - Skill 分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 技能名称，支持模糊搜索
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 * @param {string[]} [data.publishScope] 发布范围：private、organization、public
 * @param {string[]} [data.publishStatus] 发布状态：draft、publish
 */
export const getAdminSkillPageList = data => {
  return service({
    url: `${USER_API}/admin/center/skill/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - Skill 详情。
 *
 * @param {Object} data
 * @param {string} data.skillId Skill ID
 */
export const getAdminSkillDetail = data => {
  return service({
    url: `${USER_API}/admin/center/skill/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - Skill 基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.skillId Skill ID
 */
export const getAdminSkillBase = data => {
  return service({
    url: `${USER_API}/admin/center/skill/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 知识库全局分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 知识库名称
 * @param {number[]} [data.category] 类型筛选：0 文本知识库、1 问答库、2 多模态知识库
 * @param {number} [data.external] 内外部筛选：-1 全部、0 内部知识库、1 外部知识库
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 * @param {string[]} [data.publishScope] 发布范围：private、organization、public
 * @param {string[]} [data.publishStatus] 发布状态：draft、publish
 */
export const getAdminKnowledgePageList = data => {
  return service({
    url: `${USER_API}/admin/center/knowledge/page/list`,
    method: 'post',
    data,
  });
};

export const getAdminKnowledgeBase = data => {
  return service({
    url: `${USER_API}/admin/center/knowledge/base`,
    method: 'post',
    data,
  });
};

export const getAdminKnowledgeFileList = data => {
  return service({
    url: `${USER_API}/admin/center/knowledge/file/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 知识库问答对分页列表。
 * 参数同 getQaPairList。
 *
 * @param {Object} data 查询参数
 * @param {string} data.knowledgeId 知识库 ID
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 问题，支持模糊搜索
 * @param {string} [data.metaValue] 元数据值
 * @param {number[]} [data.status] 导入状态筛选：-1 全部、0 待处理、1 导入中、2 导入成功、3 导入失败
 */
export const getAdminKnowledgeQaPairList = data => {
  return service({
    url: `${USER_API}/admin/center/knowledge/qa/pair/list`,
    method: 'post',
    data,
  });
};

export const getAdminKnowledgeFileDetail = data => {
  return service({
    url: `${USER_API}/admin/center/knowledge/file/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 智能体分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 智能体名称，支持模糊搜索
 * @param {number[]} [data.category] 类别过滤：1 单智能体、2 多智能体
 * @param {string[]} [data.publishStatus] 发布状态列表
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 */
export const getAdminAssistantPageList = data => {
  return service({
    url: `${USER_API}/admin/center/assistant/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - RAG 分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] RAG 名称，支持模糊搜索
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.publishStatus] 发布状态列表
 */
export const getAdminRagPageList = data => {
  return service({
    url: `${USER_API}/admin/center/rag/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 知识问答基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.ragId 知识问答 ID
 */
export const getAdminRagBase = data => {
  return service({
    url: `${USER_API}/admin/center/rag/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 知识问答详情。
 *
 * @param {Object} data
 * @param {string} data.ragId 知识问答 ID
 */
export const getAdminRagDetail = data => {
  return service({
    url: `${USER_API}/admin/center/rag/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 智能体基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.assistantId 智能体 ID
 */
export const getAdminAssistantBase = data => {
  return service({
    url: `${USER_API}/admin/center/assistant/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 智能体详情。
 *
 * @param {Object} data
 * @param {string} data.assistantId 智能体 ID
 */
export const getAdminAssistantDetail = data => {
  return service({
    url: `${USER_API}/admin/center/assistant/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 敏感词表分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 敏感词表名称，支持模糊搜索
 * @param {string[]} [data.publishStatus] 发布状态列表
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 */
export const getAdminSensitivePageList = data => {
  return service({
    url: `${USER_API}/admin/center/sensitive/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 敏感词表基础信息（名称/备注/类型/创建人/所属组织）。
 *
 * @param {Object} data
 * @param {string} data.tableId 敏感词表 ID
 */
export const getAdminSensitiveWordBase = data => {
  return service({
    url: `${USER_API}/admin/center/sensitive/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 敏感词详情（分页词列表 + 回复设置）。
 *
 * @param {Object} data
 * @param {string} data.tableId 敏感词表 ID
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 */
export const getAdminSensitiveWordDetail = data => {
  return service({
    url: `${USER_API}/admin/center/sensitive/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - MCP 全局分页列表。
 *
 * @param {Object} data 查询参数，会作为 body.data 发送
 * @param {string} [data.name] MCP 服务名称
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.publishStatus] 发布状态列表
 * @param {string[]} [data.type] MCP 连接类型列表
 * @param {string[]} [data.userIdList] 用户 ID 列表
 */
export const getAdminMcpPageList = data => {
  return service({
    url: `${USER_API}/admin/center/mcp/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - MCP 基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.mcpId MCP ID
 * @param {string} data.type MCP 连接类型：custom 导入、server 创建
 */
export const getAdminMcpBase = data => {
  return service({
    url: `${USER_API}/admin/center/mcp/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 导入 MCP 详情（type === 'mcp'）。
 *
 * @param {Object} data
 * @param {string} data.mcpId MCP ID
 */
export const getAdminMcpCustomDetail = data => {
  return service({
    url: `${USER_API}/admin/center/mcp/custom/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 创建 MCP 详情（type === 'mcpserver'）。
 *
 * @param {Object} data
 * @param {string} data.mcpId MCP ID
 */
export const getAdminMcpServerDetail = data => {
  return service({
    url: `${USER_API}/admin/center/mcp/server/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - MCP 工具列表。
 *
 * @param {Object} data
 * @param {string} data.mcpId MCP ID
 */
export const getAdminMcpToolList = data => {
  return service({
    url: `${USER_API}/admin/center/mcp/tool/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 提示词全局分页列表。
 *
 * @param {Object} data 查询参数，会作为 body.data 发送
 * @param {string} [data.name] 提示词名称
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.publishStatus] 发布状态列表
 * @param {string[]} [data.userIdList] 用户 ID 列表
 */
export const getAdminPromptPageList = data => {
  return service({
    url: `${USER_API}/admin/center/prompt/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 提示词基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.customPromptId 提示词 ID
 */
export const getAdminPromptBase = data => {
  return service({
    url: `${USER_API}/admin/center/prompt/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 提示词详情。
 *
 * @param {Object} data
 * @param {string} data.customPromptId 提示词 ID
 */
export const getAdminPromptDetail = data => {
  return service({
    url: `${USER_API}/admin/center/prompt/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 工具全局分页列表。
 *
 * @param {Object} data 查询参数
 * @param {string} [data.name] 工具名称
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string[]} [data.publishScope] 发布范围列表
 * @param {string[]} [data.publishStatus] 发布状态列表
 * @param {string[]} [data.userIdList] 用户 ID 列表
 */
export const getAdminToolPageList = data => {
  return service({
    url: `${USER_API}/admin/center/tool/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 工具基础信息（头像/名称/描述/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.toolId 工具 ID
 * @param {string} data.type 工具类型：custom 自定义工具
 */
export const getAdminToolBase = data => {
  return service({
    url: `${USER_API}/admin/center/tool/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 工具详情（apiAuth/apiList/schema/privacyPolicy）。
 *
 * @param {Object} data
 * @param {string} data.toolId 工具 ID
 * @param {string} data.type 工具类型：custom 自定义工具
 */
export const getAdminToolDetail = data => {
  return service({
    url: `${USER_API}/admin/center/tool/detail`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 工作流分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 工作流名称
 * @param {string[]} [data.appType] 应用类型列表
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 * @param {string[]} [data.publishScope] 发布范围：private、organization、public
 * @param {string[]} [data.publishStatus] 发布状态：draft、publish
 */
export const getAdminWorkflowPageList = data => {
  return service({
    url: `${USER_API}/admin/center/workflow/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 模型分页列表。
 *
 * @param {Object} data 查询参数
 * @param {number} data.pageNo 当前页码
 * @param {number} data.pageSize 每页条数
 * @param {string} [data.name] 模型名称
 * @param {string[]} [data.modelType] 模型类型列表
 * @param {string[]} [data.orgIdList] 组织 ID 列表
 * @param {boolean} [data.isAllOrg] 是否查询全部组织
 * @param {string[]} [data.userIdList] 用户 ID 列表
 * @param {string[]} [data.publishScope] 发布范围：private、organization、public
 * @param {string[]} [data.publishStatus] 发布状态：draft、publish
 */
export const getAdminModelPageList = data => {
  return service({
    url: `${USER_API}/admin/center/model/page/list`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 模型基础信息（头像/名称/描述/发布状态/发布范围/创建人/可见用户）。
 *
 * @param {Object} data
 * @param {string} data.modelId 模型 ID
 */
export const getAdminModelBase = data => {
  return service({
    url: `${USER_API}/admin/center/model/base`,
    method: 'post',
    data,
  });
};

/**
 * 管理员中心 - 模型详情。
 *
 * @param {Object} data
 * @param {string} data.modelId 模型 ID
 */
export const getAdminModelDetail = data => {
  return service({
    url: `${USER_API}/admin/center/model/detail`,
    method: 'post',
    data,
  });
};
