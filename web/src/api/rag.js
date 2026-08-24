import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

export const getRagInfo = params => {
  return service({
    url: `${USER_API}/appspace/rag/draft`,
    method: 'get',
    params,
  });
};
export const getRagPublishedInfo = params => {
  return service({
    url: `${USER_API}/appspace/rag`,
    method: 'get',
    params,
  });
};
export const updateRag = data => {
  return service({
    url: `${USER_API}/appspace/rag`,
    method: 'put',
    data,
  });
};
export const updateRagConfig = data => {
  return service({
    url: `${USER_API}/appspace/rag/config`,
    method: 'put',
    data,
  });
};
export const createRag = data => {
  return service({
    url: `${USER_API}/appspace/rag`,
    method: 'post',
    data,
  });
};
export const delRag = data => {
  return service({
    url: `${USER_API}/appspace/rag`,
    method: 'delete',
    data,
  });
};
/**
 * 创建已发布知识问答应用的会话。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {string} data.prompt 首轮提问，用于生成会话标题。
 */
export const createRagConversation = data => {
  return service({
    url: `${USER_API}/rag/conversation`,
    method: 'post',
    data,
  });
};

/**
 * 查询已发布知识问答应用的历史会话列表。
 * @param {Object} params 查询参数。
 * @param {string} params.ragId 知识问答应用 ID。
 * @param {number} params.pageNo 页码，从 1 开始。
 * @param {number} params.pageSize 每页条数。
 * @param {string} [params.searchText] 会话标题搜索关键字。
 */
export const getRagConversationList = params => {
  return service({
    url: `${USER_API}/rag/conversation/list`,
    method: 'get',
    params,
  });
};

/**
 * 查询指定已发布知识问答会话的多轮消息详情。
 * @param {Object} params 查询参数。
 * @param {string} params.ragId 知识问答应用 ID。
 * @param {string} params.conversationId 会话 ID。
 * @param {number} params.pageNo 页码，从 1 开始。
 * @param {number} params.pageSize 每页条数。
 */
export const getRagConversationDetail = params => {
  return service({
    url: `${USER_API}/rag/conversation/detail`,
    method: 'get',
    params,
  });
};

/**
 * 删除已发布知识问答会话；传入 detailId 时仅删除指定的一轮问答。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {string} data.conversationId 会话 ID。
 * @param {string} [data.detailId] 单轮问答 ID。
 */
export const deleteRagConversation = data => {
  return service({
    url: `${USER_API}/rag/conversation`,
    method: 'delete',
    data,
  });
};

/**
 * 清空已发布知识问答会话的消息，但保留会话本身；传入 detailId 时仅清除指定的一轮问答。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {string} data.conversationId 会话 ID。
 * @param {string} [data.detailId] 单轮问答 ID。
 */
export const clearRagConversation = data => {
  return service({
    url: `${USER_API}/rag/conversation/clear`,
    method: 'delete',
    data,
  });
};

/**
 * 查询知识问答草稿态的单一会话历史。
 * @param {Object} params 查询参数。
 * @param {string} params.ragId 知识问答应用 ID。
 * @param {number} [params.pageNo=1] 页码，从 1 开始。
 * @param {number} [params.pageSize=20] 每页条数。
 */
export const getRagDraftConversationDetail = params => {
  return service({
    url: `${USER_API}/rag/conversation/draft/detail`,
    method: 'get',
    params,
  });
};

/**
 * 删除知识问答草稿态会话；传入 detailId 时仅删除指定的一轮问答。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {string} [data.detailId] 单轮问答 ID。
 */
export const deleteRagDraftConversation = data => {
  return service({
    url: `${USER_API}/rag/conversation/draft`,
    method: 'delete',
    data,
  });
};

/**
 * 查询知识问答是否存在可恢复的进行中会话。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {boolean} [data.draft=false] 是否为草稿态。
 * @param {string} [data.conversationId] 发布态会话 ID，草稿态无需传入。
 * @returns {Promise} data.hasPendingConversation 为 true 时需连接 /rag/stream/connect。
 */
export const getRagPendingConversation = data => {
  return service({
    url: `${USER_API}/rag/pending/conversation`,
    method: 'post',
    data,
  });
};

/**
 * 取消知识问答正在运行的流式会话。
 * @param {Object} data 请求参数。
 * @param {string} data.ragId 知识问答应用 ID。
 * @param {boolean} [data.draft=false] 是否为草稿态。
 * @param {string} [data.conversationId] 发布态会话 ID，草稿态无需传入。
 * @returns {Promise} 接口幂等，目标会话不存在时也视为成功。
 */
export const cancelRagStream = data => {
  return service({
    url: `${USER_API}/rag/stream/cancel`,
    method: 'post',
    data,
  });
};
