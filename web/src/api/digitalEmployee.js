import service from '@/utils/request';
import { SERVICE_API } from '@/utils/requestConstants';
import { streamSSE } from '@/api/generalAgent';

// 基础路径
const BASE_URL = `${SERVICE_API}/digital-employee`;

// ==================== 广场详情 ====================

/**
 * 数字员工广场详情
 * @param {string} employeeId - 数字员工ID
 */
export const getDigitalEmployeeDetail = params => {
  return service({
    url: `${BASE_URL}/detail`,
    method: 'get',
    params,
  });
};

// ==================== 对话管理 ====================

/**
 * 创建数字员工发布会话（绑定数字员工，创建后不可切换）
 * @param {string} employeeId - 数字员工ID（必填）
 * @param {string} title - 会话标题（必填）
 * @param {object} modelConfig - 模型配置（可选）
 */
export const createDigitalEmployeeConversation = data => {
  return service({
    url: `${BASE_URL}/conversation`,
    method: 'post',
    data,
  });
};

/**
 * 删除数字员工发布会话
 * @param {string} conversationId - 会话ID（必填）
 */
export const deleteDigitalEmployeeConversation = data => {
  return service({
    url: `${BASE_URL}/conversation`,
    method: 'delete',
    data,
  });
};

/**
 * 获取数字员工发布会话列表（按 employeeId 维度过滤）
 * @param {string} employeeId - 数字员工ID（必填）
 * @param {number} pageNo - 页码（必填）
 * @param {number} pageSize - 每页数量（必填）
 * @param {string} searchText - 标题关键词（可选）
 */
export const getDigitalEmployeeConversationList = params => {
  return service({
    url: `${BASE_URL}/conversation/list`,
    method: 'get',
    params,
  });
};

/**
 * 获取会话详情（含历史消息，ES 回放，格式同 wga）
 * @param {string} conversationId - 会话ID
 * @param {number} pageNo - 页码，默认1
 * @param {number} pageSize - 每页数量，默认1000
 */
export const getDigitalEmployeeConversationDetail = params => {
  return service({
    url: `${BASE_URL}/conversation/detail`,
    method: 'get',
    params,
  });
};

// ==================== 会话配置 ====================

/**
 * 获取数字员工发布会话配置
 * 返回：conversationId, employeeId, title, modelConfig
 * @param {string} conversationId - 会话ID（必填）
 */
export const getDigitalEmployeeConversationConfig = params => {
  return service({
    url: `${BASE_URL}/conversation/config`,
    method: 'get',
    params,
  });
};

/**
 * 修改数字员工发布会话配置（仅模型配置）
 * @param {string} conversationId - 会话ID（必填）
 * @param {object} modelConfig - 模型配置 { modelId, model, provider, displayName, modelType, config }（必填）
 */
export const updateDigitalEmployeeConversationConfig = data => {
  return service({
    url: `${BASE_URL}/conversation/config`,
    method: 'put',
    data,
  });
};

// ==================== SSE 对话 ====================

/**
 * 数字员工发布对话（SSE 流式，wga 模式，固定 DIP Agent）
 * 两步模式：会话须先通过 createDigitalEmployeeConversation 创建，conversationId 必填；
 * 模型配置由后端从会话配置读回，对话时不随请求发送。
 * 文件通过 messages[].content 的 binary 项传递（同通用智能体）。
 * @param {string} employeeId - 数字员工ID（必填）
 * @param {string} conversationId - 会话ID（必填，须先创建）
 * @param {Array} messages - 消息列表 [{ id, role, content }]（必填）
 * @param {function} onMessage - 消息回调
 * @param {function} onError - 错误回调
 * @param {function} onOpen - 连接建立回调
 * @param {AbortSignal} signal - 取消信号
 * @param {number} timeout - 超时时间（毫秒），默认 5 分钟
 */
export const chatDigitalEmployeeConversation = async ({
  employeeId,
  conversationId,
  messages,
  onMessage,
  onError,
  onOpen,
  signal,
  timeout,
}) => {
  return streamSSE({
    url: `${location.origin}${BASE_URL}/chat`,
    body: {
      employeeId,
      conversationId,
      messages,
    },
    onMessage,
    onError,
    onOpen,
    signal,
    timeout,
  });
};
