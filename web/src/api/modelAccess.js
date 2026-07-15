import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取列表
export const fetchModelList = params => {
  return service({
    url: `${USER_API}/model/list`,
    method: 'get',
    params,
  });
};

// 获取单个模型
export const getModelDetail = params => {
  return service({
    url: `${USER_API}/model`,
    method: 'get',
    params,
  });
};

// 获取模型供应商列表
export const fetchProviderList = params => {
  return service({
    url: `${USER_API}/model/import/providers`,
    method: 'get',
    params,
  });
};

// 支持深度思考开关连接测试接口
export const testLinkThinking = data => {
  return service({
    url: `${USER_API}/model/validate-thinking`,
    method: 'post',
    data,
  });
};

// 创建
export const addModel = data => {
  return service({
    url: `${USER_API}/model`,
    method: 'post',
    data,
  });
};
// 编辑
export const editModel = data => {
  return service({
    url: `${USER_API}/model`,
    method: 'put',
    data,
  });
};
// 删除
export const deleteModel = data => {
  return service({
    url: `${USER_API}/model`,
    method: 'delete',
    data,
  });
};
// 修改状态
export const changeModelStatus = data => {
  return service({
    url: `${USER_API}/model/status`,
    method: 'put',
    data,
  });
};

//获取embedding列表
export const getEmbeddingList = params => {
  return service({
    url: `${USER_API}/model/select/embedding`,
    method: 'get',
    params,
  });
};

//获取多模态embedding列表
export const getMultiEmbeddingList = params => {
  return service({
    url: `${USER_API}/model/select/multi-embedding`,
    method: 'get',
    params,
  });
};

//获取rerank模型列表
export const getRerankList = () => {
  return service({
    url: `${USER_API}/model/select/rerank`,
    method: 'get',
  });
};

//获取多模态rerank模型列表
export const getMultiRerankList = () => {
  return service({
    url: `${USER_API}/model/select/multi-rerank`,
    method: 'get',
  });
};

//获取选择模型下拉列表
export const selectModelList = () => {
  return service({
    url: `${USER_API}/model/select/llm`,
    method: 'get',
  });
};

//获取ASR模型列表
export const selectASRList = () => {
  return service({
    url: `${USER_API}/model/select/sync-asr`,
    method: 'get',
  });
};

// 获取模型ID列表
export const fetchModelIdList = params => {
  return service({
    url: `${USER_API}/model/recommend`,
    method: 'get',
    params,
  });
};
