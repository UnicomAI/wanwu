import service from '@/utils/request';
import { USER_API, OPENURL_API } from '@/utils/requestConstants';

export const createAgent = data => {
  return service({
    url: `${USER_API}/assistant`,
    method: 'post',
    data,
  });
};

export const updateAgent = data => {
  return service({
    url: `${USER_API}/assistant`,
    method: 'put',
    data,
  });
};
export const delAgent = data => {
  return service({
    url: `${USER_API}/assistant`,
    method: 'delete',
    data,
  });
};
export const getAgentPublishedInfo = params => {
  return service({
    url: `${USER_API}/assistant`,
    method: 'get',
    params,
  });
};

export const getMultiAgentList = params => {
  return service({
    url: `${USER_API}/assistant/select`,
    method: 'get',
    params,
  });
};
export const bindMultiAgent = data => {
  return service({
    url: `${USER_API}/assistant/multi-agent`,
    method: 'post',
    data,
  });
};
export const unbindMultiAgent = data => {
  return service({
    url: `${USER_API}/assistant/multi-agent`,
    method: 'delete',
    data,
  });
};
export const updateMultiAgent = data => {
  return service({
    url: `${USER_API}/assistant/multi-agent/config`,
    method: 'put',
    data,
  });
};
export const switchMultiAgent = data => {
  return service({
    url: `${USER_API}/assistant/multi-agent/switch`,
    method: 'put',
    data,
  });
};

export const getAgentInfo = params => {
  return service({
    url: `${USER_API}/assistant/draft`,
    method: 'get',
    params,
  });
};
export const putAgentInfo = data => {
  return service({
    url: `${USER_API}/assistant/config`,
    method: 'put',
    data,
  });
};
export const createConversation = data => {
  return service({
    url: `${USER_API}/assistant/conversation`,
    method: 'post',
    data,
  });
};
export const delConversation = data => {
  return service({
    url: `${USER_API}/assistant/conversation`,
    method: 'delete',
    data,
  });
};
export const getConversationHistory = params => {
  return service({
    url: `${USER_API}/assistant/conversation/detail`,
    method: 'get',
    params,
  });
};
export const getConversationlist = params => {
  return service({
    url: `${USER_API}/assistant/conversation/list`,
    method: 'get',
    params,
  });
};
export const getActionInfo = params => {
  return service({
    url: `${USER_API}/assistant/action`,
    method: 'get',
    params,
  });
};
export const editActionInfo = data => {
  return service({
    url: `${USER_API}/assistant/action`,
    method: 'put',
    data,
  });
};
export const addActionInfo = data => {
  return service({
    url: `${USER_API}/assistant/action`,
    method: 'post',
    data,
  });
};
export const delActionInfo = data => {
  return service({
    url: `${USER_API}/assistant/action`,
    method: 'delete',
    data,
  });
};
export const enableAction = data => {
  return service({
    url: `${USER_API}/assistant/action/enable`,
    method: 'put',
    data,
  });
};
export const addWorkFlowInfo = data => {
  return service({
    url: `${USER_API}/assistant/tool/workflow`,
    method: 'post',
    data,
  });
};
export const delWorkFlowInfo = data => {
  return service({
    url: `${USER_API}/assistant/tool/workflow`,
    method: 'delete',
    data,
  });
};
export const enableWorkFlow = data => {
  return service({
    url: `${USER_API}/assistant/tool/workflow/switch`,
    method: 'put',
    data,
  });
};
export const agentStream = data => {
  return service({
    url: `${USER_API}/assistant/stream`,
    method: 'post',
    data,
  });
};
export const agentTestStream = data => {
  return service({
    url: `${USER_API}/assistant/test/stream`,
    method: 'post',
    data,
  });
};
export const getAgentList = params => {
  return service({
    url: `${USER_API}/assistant/list`,
    method: 'get',
    params,
  });
};

//删除mcp工具
export const deleteMcp = data => {
  return service({
    url: `${USER_API}/assistant/tool/mcp`,
    method: 'delete',
    data,
  });
};
//添加mcp工具
export const addMcp = data => {
  return service({
    url: `${USER_API}/assistant/tool/mcp`,
    method: 'post',
    data,
  });
};
//启停mcp工具
export const enableMcp = data => {
  return service({
    url: `${USER_API}/assistant/tool/mcp/switch`,
    method: 'put',
    data,
  });
};

// 添加skill
export const addSkill = data => {
  return service({
    url: `${USER_API}/assistant/skill`,
    method: 'post',
    data,
  });
};

//启停skill
export const enableSkill = data => {
  return service({
    url: `${USER_API}/assistant/skill/switch`,
    method: 'put',
    data,
  });
};

// 删除skill
export const deleteSkill = data => {
  return service({
    url: `${USER_API}/assistant/skill`,
    method: 'delete',
    data,
  });
};

// 删除自定义、内置工具
export const delCustomBuiltIn = data => {
  return service({
    url: `${USER_API}/assistant/tool`,
    method: 'delete',
    data,
  });
};
// 添加自定义、内置工具
export const addCustomBuiltIn = data => {
  return service({
    url: `${USER_API}/assistant/tool`,
    method: 'post',
    data,
  });
};
// 启停自定义、内置工具
export const switchCustomBuiltIn = data => {
  return service({
    url: `${USER_API}/assistant/tool/switch`,
    method: 'put',
    data,
  });
};
//工具列表
export const toolList = data => {
  return service({
    url: `${USER_API}/tool/select`,
    method: 'get',
    params: data,
  });
};
//工具下面的action列表
export const toolActionList = data => {
  return service({
    url: `${USER_API}/tool/action/list`,
    method: 'get',
    params: data,
  });
};
//内置工具下面的action详情
export const toolActionDetail = data => {
  return service({
    url: `${USER_API}/tool/action/detail`,
    method: 'get',
    params: data,
  });
};
//mcp工具列表
export const mcptoolList = data => {
  return service({
    url: `${USER_API}/mcp/select`,
    method: 'get',
    params: data,
  });
};
//mcp工具下面的action列表
export const mcpActionList = data => {
  return service({
    url: `${USER_API}/mcp/action/list`,
    method: 'get',
    params: data,
  });
};

//编辑url
export const editOpenurl = data => {
  return service({
    url: `${USER_API}/appspace/app/openurl`,
    method: 'put',
    data,
  });
};
//创建url
export const createOpenurl = data => {
  return service({
    url: `${USER_API}/appspace/app/openurl`,
    method: 'post',
    data,
  });
};
//删除应用url
export const delOpenurl = data => {
  return service({
    url: `${USER_API}/appspace/app/openurl`,
    method: 'delete',
    data,
  });
};
//获取应用url列表
export const getOpenurl = data => {
  return service({
    url: `${USER_API}/appspace/app/openurl/list`,
    method: 'get',
    params: data,
  });
};
//启停应用url状态
export const switchOpenurl = data => {
  return service({
    url: `${USER_API}/appspace/app/openurl/status`,
    method: 'put',
    data,
  });
};

//获取智能体openurl信息
export const getOpenurlInfo = (suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}`,
    method: 'get',
    ...config,
    isOpenUrl: true,
  });
};
//智能体openurl创建智能体对话
export const openurlConversation = (data, suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/conversation`,
    method: 'post',
    data,
    ...config,
    isOpenUrl: true,
  });
};
//删除智能体openurl创建智能体对话
export const delOpenurlConversation = (data, suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/conversation`,
    method: 'delete',
    data,
    ...config,
    isOpenUrl: true,
  });
};
//智能体openurl详情历史列表
export const OpenurlConverHistory = (data, suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/conversation/detail`,
    method: 'get',
    params: data,
    ...config,
    isOpenUrl: true,
  });
};
//智能体openurl对话列表
export const OpenurlConverList = (suffix, params = {}, config = {}) => {
  const isOldConfig = params && (params.headers || params.isOpenUrl);
  const requestParams = isOldConfig ? {} : params;
  const requestConfig = isOldConfig ? params : config;
  return service({
    url: `${OPENURL_API}/agent/${suffix}/conversation/list`,
    method: 'get',
    params: requestParams,
    ...requestConfig,
    isOpenUrl: true,
  });
};
//智能体openurl流式对话
export const OpenurlStream = (data, suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/stream`,
    method: 'post',
    data,
    ...config,
    isOpenUrl: true,
  });
};

//智能体openurl删除会话
export const openurlConverDel = (data, suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/conversation/clear`,
    method: 'delete',
    data,
    ...config,
    isOpenUrl: true,
  });
};
//更新博查rerank模型
export const updateRerank = data => {
  return service({
    url: `${USER_API}/assistant/tool/config`,
    method: 'put',
    data,
  });
};
//智能体工具-工作流列表
export const getWorkflowList = data => {
  return service({
    url: `${USER_API}/workflow/select`,
    method: 'get',
    params: data,
  });
};

/**
 * 获取推荐问题的接口URL
 * 根据类型和助手ID，返回对应环境下的推荐问题接口地址
 *
 * @param {string} [type='agentChat']
 *   - 'agentChat': 默认场景
 *   - 'webChat': openurl场景
 * @param {string} [assistantId] - agentID（在非agentChat类型时必填）
 */
export const getRecommendQuestionUrl = (type = 'agentChat', assistantId) => {
  if (type === 'agentChat') {
    return `${USER_API}/assistant/question/recommend`;
  } else {
    // 非agentChat类型需要assistantId
    if (!assistantId) {
      throw new Error('非agentChat类型必须提供assistantId参数');
    }
    return `${OPENURL_API}/agent/${assistantId}/recommend`;
  }
};

// 获取草稿智能体会话历史
export const getConversationDraftHistory = params => {
  return service({
    url: `${USER_API}/assistant/conversation/draft/detail`,
    method: 'get',
    params,
  });
};

// 清空/删除草稿智能会话
export const delConversationDraft = data => {
  return service({
    url: `${USER_API}/assistant/conversation/draft`,
    method: 'delete',
    data,
  });
};

// 清空对话
export const clearConversation = data => {
  return service({
    url: `${USER_API}/assistant/conversation/clear`,
    method: 'delete',
    data,
  });
};

/**
 * 获取进行中的智能体会话详情
 * @param {Object} params - 请求参数
 * @param {string} params.assistantId - 智能体助手ID
 * @param {string} params.conversationId - 会话ID
 * @param {boolean} [params.draft] - 是否为草稿模式
 */
export const getPendingConversation = (data, config = {}) => {
  return service({
    url: `${USER_API}/assistant/pending/conversation`,
    method: 'post',
    data,
    ...config,
  });
};

/**
 * 智能体流式问答手动停止
 * @param {Object} data - 请求数据
 * @param {string} data.assistantId - 智能体助手ID
 * @param {string} data.conversationId - 会话ID
 * @param {boolean} [params.draft] - 是否为草稿模式
 */
export const cancelAgentStream = (data, config = {}) => {
  return service({
    url: `${USER_API}/assistant/stream/cancel`,
    method: 'post',
    data,
    ...config,
  });
};

/**
 * 获取进行中的智能体会话详情（weburl）
 * @param {Object} params - 请求参数
 * @param {string} params.conversationId - 会话ID
 */
export const getOpenurlPendingConversation = (suffix, params, config) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/pending/conversation`,
    method: 'get',
    params,
    ...config,
    isOpenUrl: true,
  });
};

/**
 * 智能体流式问答手动停止（weburl）
 * @param {Object} data - 请求数据
 * @param {string} data.conversationId - 会话ID
 */
export const cancelOpenurlAgentStream = (suffix, data, config) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/stream/cancel`,
    method: 'post',
    data,
    ...config,
    isOpenUrl: true,
  });
};

// 获取智能体openurl模型配置
export const getOpenurlAgentLlm = (suffix, config = {}) => {
  return service({
    url: `${OPENURL_API}/agent/${suffix}/llm`,
    method: 'get',
    ...config,
    isOpenUrl: true,
  });
};
