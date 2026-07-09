import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

/**
 * 模型统计接口
 */

// 获取模型下拉列表
export const getModelSelect = data => {
  return service({
    url: `${USER_API}/statistic/model/select`,
    method: 'post',
    data,
  });
};

// 获取模型统计数据
export const getModelData = data => {
  return service({
    url: `${USER_API}/statistic/model`,
    method: 'post',
    data,
  });
};

// 获取模型列表
export const fetchModelList = data => {
  const type = data.type;
  delete data.type;
  return service({
    url: `${USER_API}/statistic/model/${type || 'list'}`,
    method: 'post',
    data,
  });
};

// 模型数据导出
export const exportModelData = (data, type) => {
  return service({
    url: `${USER_API}/statistic/model/${type || 'list'}/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取模型用户统计列表
export const fetchModelUserList = data => {
  return service({
    url: `${USER_API}/statistic/model/user/list`,
    method: 'post',
    data,
  });
};

// 导出模型用户统计数据
export const exportModelUserData = data => {
  return service({
    url: `${USER_API}/statistic/model/user/export`,
    method: 'post',
    data,
    responseType: 'blob',
  });
};

// 获取模型应用统计列表
export const fetchModelAppList = data => {
  return service({
    url: `${USER_API}/statistic/model/app/list`,
    method: 'post',
    data,
  });
};

// 导出模型应用统计数据
export const exportModelAppData = data => {
  return service({
    url: `${USER_API}/statistic/model/app/export`,
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
    url: `${USER_API}/statistic/app/select`,
    method: 'post',
    data,
  });
};

// 获取应用统计数据
export const getAppData = data => {
  return service({
    url: `${USER_API}/statistic/app`,
    method: 'post',
    data,
  });
};

// 获取应用统计列表
export const fetchAppList = data => {
  return service({
    url: `${USER_API}/statistic/app/list`,
    method: 'post',
    data,
  });
};

// 应用数据导出
export const exportAppData = data => {
  return service({
    url: `${USER_API}/statistic/app/export`,
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
    url: `${USER_API}/statistic/api/select`,
    method: 'post',
    data,
  });
};

// 获取API路径列表
export const getApiRoutes = params => {
  return service({
    url: `${USER_API}/statistic/api/routes`,
    method: 'get',
    params,
  });
};

// 获取API统计数据
export const getApiData = data => {
  return service({
    url: `${USER_API}/statistic/api`,
    method: 'post',
    data,
  });
};

// 获取API列表
export const fetchApiList = data => {
  const type = data.type;
  delete data.type;
  return service({
    url: `${USER_API}/statistic/api/${type || 'list'}`,
    method: 'post',
    data,
  });
};

// API数据导出
export const exportApiData = (data, type) => {
  return service({
    url: `${USER_API}/statistic/api/${type || 'list'}/export`,
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
