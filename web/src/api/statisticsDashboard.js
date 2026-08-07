import service from '@/utils/request';
import { USER_API, USER_API_V2 } from '@/utils/requestConstants';

/**
 * 模型统计接口
 */

// 获取模型下拉列表
export const getModelSelect = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/select`,
    method: 'post',
    data,
  });
};

// 获取模型统计数据概览
export const getModelData = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/overview`,
    method: 'post',
    data,
  });
};

// 获取模型统计图表数据+排行
export const getModelChart = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/chart`,
    method: 'post',
    data,
  });
};

// 获取模型列表
export const fetchModelList = data => {
  const type = data.type;
  delete data.type;
  const path = type === 'list' ? '' : `${type}/`;
  return service({
    url: `${USER_API_V2}/statistic/model/${path}/list`,
    method: 'post',
    data,
  });
};

// 模型数据导出
export const exportModelData = (data, type) => {
  return service({
    url: `${USER_API_V2}/statistic/model/${type || 'list'}/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取模型用户统计列表
export const fetchModelUserList = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/list/user`,
    method: 'post',
    data,
  });
};

// 导出模型用户统计数据
export const exportModelUserData = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/list/user/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取模型应用统计列表
export const fetchModelAppList = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/list/app`,
    method: 'post',
    data,
  });
};

// 导出模型应用统计数据
export const exportModelAppData = data => {
  return service({
    url: `${USER_API_V2}/statistic/model/list/app/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

/**
 * 应用统计接口
 */

// 获取应用下拉列表
export const getAppSelect = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/select`,
    method: 'post',
    data,
  });
};

// 获取应用统计数据概览
export const getAppData = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/overview`,
    method: 'post',
    data,
  });
};

// 获取应用统计图表数据+排行
export const getAppChart = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/chart`,
    method: 'post',
    data,
  });
};

// 获取应用统计列表
export const fetchAppList = data => {
  const type = data.type;
  delete data.type;
  const path = type === 'list' ? '' : `${type}/`;
  return service({
    url: `${USER_API_V2}/statistic/app/${path}list`,
    method: 'post',
    data,
  });
};

// 应用数据导出
export const exportAppData = (data, type) => {
  return service({
    url: `${USER_API_V2}/statistic/app/${type || 'list'}/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取应用用户统计列表
export const fetchAppUserList = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/list/user`,
    method: 'post',
    data,
  });
};

// 导出应用用户统计数据
export const exportAppUserData = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/list/user/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取应用模型统计列表
export const fetchAppModelList = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/list/model`,
    method: 'post',
    data,
  });
};

// 导出应用模型统计数据
export const exportAppModelData = data => {
  return service({
    url: `${USER_API_V2}/statistic/app/list/model/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

/**
 * API统计接口
 */

// 获取API下拉列表
export const getApiSelect = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/select`,
    method: 'post',
    data,
  });
};

// 获取API路径列表
export const getApiRoutes = params => {
  return service({
    url: `${USER_API_V2}/statistic/api/routes`,
    method: 'get',
    params,
  });
};

// 获取API统计数据概览
export const getApiData = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/overview`,
    method: 'post',
    data,
  });
};

// 获取API统计图表数据+排行
export const getApiChart = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/chart`,
    method: 'post',
    data,
  });
};

// 获取API列表
export const fetchApiList = data => {
  const type = data.type;
  delete data.type;
  const path = type === 'list' ? '' : `${type}/`;
  return service({
    url: `${USER_API_V2}/statistic/api/${path}list`,
    method: 'post',
    data,
  });
};

// API数据导出
export const exportApiData = (data, type) => {
  return service({
    url: `${USER_API_V2}/statistic/api/${type || 'list'}/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取API应用统计列表
export const fetchApiAppList = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/list/app`,
    method: 'post',
    data,
  });
};

// 导出API应用统计数据
export const exportApiAppData = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/list/app/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取API模型统计列表
export const fetchApiModelList = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/list/model`,
    method: 'post',
    data,
  });
};

// 导出API模型统计数据
export const exportApiModelData = data => {
  return service({
    url: `${USER_API_V2}/statistic/api/list/model/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

/**
 * 全局组织和用户接口
 */

// 获取组织列表
export const fetchOrgs = params => {
  return service({
    url: `${USER_API}/statistic/orgs/select`,
    method: 'get',
    params,
  });
};

// 获取用户列表
export const fetchUsers = params => {
  return service({
    url: `${USER_API}/statistic/users/select`,
    method: 'get',
    params,
  });
};
