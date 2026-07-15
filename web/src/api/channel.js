import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取扫码二维码
export const fetchScanQrCode = channelType => {
  return service({
    url: `${USER_API}/channel/qrcode/${channelType}`,
    method: 'post',
  });
};

// 轮询扫码状态
export const pollScanStatus = (channelType, sessionId) => {
  return service({
    url: `${USER_API}/channel/qrcode/${channelType}/status/${sessionId}`,
    method: 'get',
  });
};

// 结束轮询请求接口
export const finishPollScan = (channelType, sessionId) => {
  return service({
    url: `${USER_API}/channel/qrcode/${channelType}/${sessionId}`,
    method: 'delete',
  });
};

// 获取渠道列表
export const fetchChannelList = params => {
  return service({
    url: `${USER_API}/channel/channels`,
    method: 'get',
    params,
  });
};

// 创建渠道
export const createChannel = data => {
  return service({
    url: `${USER_API}/channel/channels`,
    method: 'post',
    data,
  });
};

// 编辑渠道
export const editChannel = (id, data) => {
  return service({
    url: `${USER_API}/channel/channels/${id}`,
    method: 'put',
    data,
  });
};

// 删除渠道
export const deleteChannel = (id, data) => {
  return service({
    url: `${USER_API}/channel/channels/${id}`,
    method: 'delete',
    data,
  });
};

// 渠道开关
export const changeChannelStatus = (id, data) => {
  return service({
    url: `${USER_API}/channel/channels/${id}/status`,
    method: 'post',
    data,
  });
};

// 断开连接渠道
export const disconnectChannel = (id, data) => {
  return service({
    url: `${USER_API}/channel/channels/${id}/disconnect`,
    method: 'post',
    data,
  });
};

// 获取API下拉列表
export const getApiSelect = data => {
  return service({
    url: `${USER_API}/channel/apikeys`,
    method: 'get',
    data,
  });
};

// 获取应用下拉列表
export const getAppSelect = appType => {
  return service({
    url: `${USER_API}/channel/${appType}`,
    method: 'get',
  });
};

// 获取模型下拉列表
export const getModelSelect = params => {
  return service({
    url: `${USER_API}/channel/models`,
    method: 'get',
    params: {
      modelType: 'llm',
      ...params,
    },
  });
};

// 获取场景下拉列表
export const getSceneSelect = params => {
  return service({
    url: `${USER_API}/channel/wga/sub-agents`,
    method: 'get',
    params,
  });
};

// 获取数字员工下拉列表
export const getEmployeeSelect = params => {
  return service({
    url: `${USER_API}/channel/dip/employees`,
    method: 'get',
    params,
  });
};
