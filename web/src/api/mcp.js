import request from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

/*----自定义工具------*/
export const getCustomList = data => {
  return request({
    url: `${USER_API}/tool/custom/list`,
    method: 'get',
    params: data,
  });
};

export const getCustom = data => {
  return request({
    url: `${USER_API}/tool/custom`,
    method: 'get',
    params: data,
  });
};

export const editCustom = data => {
  return request({
    url: `${USER_API}/tool/custom`,
    method: 'put',
    data: data,
  });
};

export const addCustom = data => {
  return request({
    url: `${USER_API}/tool/custom`,
    method: 'post',
    data: data,
  });
};

export const deleteCustom = data => {
  return request({
    url: `${USER_API}/tool/custom`,
    method: 'delete',
    data: data,
  });
};

export const getSchema = data => {
  return request({
    url: `${USER_API}/tool/custom/schema`,
    method: 'post',
    data: data,
  });
};

/*---创建mcp---*/
export const getServerList = data => {
  return request({
    url: `${USER_API}/mcp/server/list`,
    method: 'get',
    params: data,
  });
};

export const getServer = data => {
  return request({
    url: `${USER_API}/mcp/server`,
    method: 'get',
    params: data,
  });
};

export const addServer = data => {
  return request({
    url: `${USER_API}/mcp/server`,
    method: 'post',
    data: data,
  });
};

export const editServer = data => {
  return request({
    url: `${USER_API}/mcp/server`,
    method: 'put',
    data: data,
  });
};

export const deleteServer = data => {
  return request({
    url: `${USER_API}/mcp/server`,
    method: 'delete',
    data: data,
  });
};

export const addServerTool = data => {
  return request({
    url: `${USER_API}/mcp/server/tool`,
    method: 'post',
    data: data,
  });
};

export const editServerTool = data => {
  return request({
    url: `${USER_API}/mcp/server/tool`,
    method: 'put',
    data: data,
  });
};

export const deleteServerTool = data => {
  return request({
    url: `${USER_API}/mcp/server/tool`,
    method: 'delete',
    data: data,
  });
};

export const addOpenapi = data => {
  return request({
    url: `${USER_API}/mcp/server/tool/openapi`,
    method: 'post',
    data: data,
  });
};

/*---导入mcp---*/
export const getList = data => {
  return request({
    url: `${USER_API}/mcp/list`,
    method: 'get',
    params: data,
  });
};

export const getDetail = data => {
  return request({
    url: `${USER_API}/mcp`,
    method: 'get',
    params: data,
  });
};

export const setDelete = data => {
  return request({
    url: `${USER_API}/mcp`,
    method: 'delete',
    data: data,
  });
};

export const getTools = data => {
  return request({
    url: `${USER_API}/mcp/tool/list`,
    method: 'post',
    data,
  });
};

export const setCreate = data => {
  return request({
    url: `${USER_API}/mcp`,
    method: 'post',
    data: data,
  });
};

export const setUpdate = data => {
  return request({
    url: `${USER_API}/mcp`,
    method: 'put',
    data: data,
  });
};

/*---第三方MCP广场---*/
export const getPublicMcpList = data => {
  return request({
    url: `${USER_API}/mcp/square/list`,
    method: 'get',
    params: data,
  });
};
export const getPublicMcpInfo = data => {
  return request({
    url: `${USER_API}/mcp/square`,
    method: 'get',
    params: data,
  });
};
export const getRecommendsList = data => {
  return request({
    url: `${USER_API}/mcp/square/recommend`,
    method: 'get',
    params: data,
  });
};

/*----内置工具------*/
export const getBuiltInList = data => {
  return request({
    url: `${USER_API}/tool/square/list`,
    method: 'get',
    params: data,
  });
};

export const getToolDetail = data => {
  return request({
    url: `${USER_API}/tool/square`,
    method: 'get',
    params: data,
  });
};

export const changeApiKey = data => {
  return request({
    url: `${USER_API}/tool/builtin`,
    method: 'post',
    data,
  });
};

export const getOAuthProviders = () => request({ url: `${USER_API}/mcp/oauth/providers`, method: 'get' });
export const mcpOAuthAction = (action, data) => request({ url: `${USER_API}/mcp/oauth/${action}`, method: 'post', data });
