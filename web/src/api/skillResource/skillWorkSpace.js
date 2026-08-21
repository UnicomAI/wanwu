/*---Skill Workspace---*/
import request from '@/utils/request';
import { SERVICE_API } from '@/utils/requestConstants';
// 获取工作区文件列表
export const getSkillWorkspaceFiles = customSkillId => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/files`,
    method: 'get',
    params: { customSkillId },
  });
};

// 读取文件内容
export const getSkillWorkspaceFile = (customSkillId, path) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file`,
    method: 'get',
    params: { customSkillId, path },
  });
};

// 获取 skill 内容文件列表（管理员中心只读概览）
export const getSkillContentFiles = customSkillId => {
  return request({
    url: `${SERVICE_API}/agent/skill/content/files`,
    method: 'get',
    params: { customSkillId },
  });
};

// 读取 skill 内容文件（管理员中心只读概览）
export const getSkillContentFile = (customSkillId, path) => {
  return request({
    url: `${SERVICE_API}/agent/skill/content/file`,
    method: 'get',
    params: { customSkillId, path },
  });
};

// 下载文件或目录
export const downloadSkillWorkspace = (customSkillId, path) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/download`,
    method: 'get',
    params: { customSkillId, path },
    responseType: 'blob',
  });
};

// 保存文件内容
export const updateSkillWorkspaceFile = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file`,
    method: 'put',
    data: { ...data, customSkillId },
  });
};

// 搜索文件内容
export const searchSkillWorkspace = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/search`,
    method: 'post',
    data: { ...data, customSkillId },
  });
};

/*---Skill Workspace Git---*/
// 获取 Git 提交历史
export const getSkillWorkspaceGitLog = (customSkillId, params) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/log`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 获取 Git diff
export const getSkillWorkspaceGitDiff = (customSkillId, params) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/diff`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 获取 Git 历史文件内容
export const getSkillWorkspaceGitFile = (customSkillId, params) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/file`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 获取 Git 单文件 diff
export const getSkillWorkspaceGitFileDiff = (customSkillId, params) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/file-diff`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 获取 Git 工作区状态
export const getSkillWorkspaceGitStatus = customSkillId => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/status`,
    method: 'get',
    params: { customSkillId },
  });
};

// 暂存文件
export const postSkillWorkspaceGitAdd = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/add`,
    method: 'post',
    data: { ...data, customSkillId },
  });
};

// 取消暂存文件
export const postSkillWorkspaceGitReset = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/reset`,
    method: 'post',
    data: { ...data, customSkillId },
  });
};

// 提交已暂存的变更
export const postSkillWorkspaceGitCommit = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/commit`,
    method: 'post',
    data: { ...data, customSkillId },
  });
};

// 获取工作区未暂存 diff
export const getSkillWorkspaceGitDiffWorking = (customSkillId, params = {}) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/diff-working`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 获取已暂存 diff
export const getSkillWorkspaceGitDiffStaged = (customSkillId, params = {}) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/diff-staged`,
    method: 'get',
    params: { ...params, customSkillId },
  });
};

// 删除文件/目录
export const deleteSkillWorkspaceFile = (customSkillId, path) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file`,
    method: 'delete',
    params: { customSkillId, path },
  });
};

// 放弃未暂存更改
export const postSkillWorkspaceGitDiscard = (customSkillId, data) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/git/discard`,
    method: 'post',
    data: { ...data, customSkillId },
  });
};
