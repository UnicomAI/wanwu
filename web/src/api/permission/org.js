import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

// 获取组织列表
export const fetchOrgList = params => {
  return service({
    url: `${USER_API}/org/list`,
    method: 'get',
    params,
  });
};
// 获取组织详情
export const fetchOrgDetail = params => {
  return service({
    url: `${USER_API}/org/info`,
    method: 'get',
    params,
  });
};
// 创建组织
export const createOrg = data => {
  return service({
    url: `${USER_API}/org`,
    method: 'post',
    data,
  });
};
// 编辑组织
export const editOrg = data => {
  return service({
    url: `${USER_API}/org`,
    method: 'put',
    data,
  });
};
// 删除组织
export const deleteOrg = data => {
  return service({
    url: `${USER_API}/org`,
    method: 'delete',
    data,
  });
};
// 修改组织状态
export const changeOrgStatus = data => {
  return service({
    url: `${USER_API}/org/status`,
    method: 'put',
    data,
  });
};
// 获取组织管理左侧组织树
export const fetchOrgTree = () => {
  return service({
    url: `${USER_API}/org/tree`,
    method: 'get',
  });
};
// 获取导航组织列表
export const fetchOrgs = () => {
  return service({
    url: `${USER_API}/org/select`,
    method: 'get',
  });
};
