import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取用户列表
export const fetchUserList = params => {
  return service({
    url: `${USER_API}/user/list`,
    method: 'get',
    params,
  });
};

// 获取角色列表用户
export const fetchRoleList = params => {
  return service({
    url: `${USER_API}/role/select`,
    method: 'get',
    params,
  });
};
// 创建用户
export const createUser = data => {
  return service({
    url: `${USER_API}/user`,
    method: 'post',
    data,
  });
};
// 批量新增
export const batchCreateUser = (data, config) => {
  return service({
    url: `${USER_API}/user/batch`,
    method: 'post',
    data,
    config,
  });
};
// 编辑用户
export const editUser = data => {
  return service({
    url: `${USER_API}/user`,
    method: 'put',
    data,
  });
};
// 删除用户
export const deleteUser = data => {
  return service({
    url: `${USER_API}/user`,
    method: 'delete',
    data,
  });
};
// 修改用户状态
export const changeUserStatus = data => {
  return service({
    url: `${USER_API}/user/status`,
    method: 'put',
    data,
  });
};
// 获取邀请用户时的用户列表
export const fetchInviteUser = params => {
  return service({
    url: `${USER_API}/org/other/select`,
    method: 'get',
    params,
  });
};
// 邀请用户
export const inviteUser = data => {
  return service({
    url: `${USER_API}/org/user`,
    method: 'post',
    data,
  });
};
