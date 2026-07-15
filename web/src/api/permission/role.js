import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取角色列表
export const fetchRoleList = params => {
  return service({
    url: `${USER_API}/role/list`,
    method: 'get',
    params,
  });
};
// 获取角色详情
export const fetchRoleDetail = params => {
  return service({
    url: `${USER_API}/role/info`,
    method: 'get',
    params,
  });
};
// 创建角色
export const createRole = data => {
  return service({
    url: `${USER_API}/role`,
    method: 'post',
    data,
  });
};
// 编辑角色
export const editRole = data => {
  return service({
    url: `${USER_API}/role`,
    method: 'put',
    data,
  });
};
// 删除角色
export const deleteRole = data => {
  return service({
    url: `${USER_API}/role`,
    method: 'delete',
    data,
  });
};
// 修改角色状态
export const changeRoleStatus = data => {
  return service({
    url: `${USER_API}/role/status`,
    method: 'put',
    data,
  });
};
// 获取关联用户
export const fetchRoleUsers = params => {
  return service({
    url: `${USER_API}/role/users`,
    method: 'get',
    params,
  });
};
// 移除关联用户
export const removeRoleUser = data => {
  return service({
    url: `${USER_API}/role/user`,
    method: 'delete',
    data,
  });
};
// 获取权限树
export const fetchPermTree = params => {
  return service({
    url: `${USER_API}/role/template`,
    method: 'get',
    params,
  });
};
