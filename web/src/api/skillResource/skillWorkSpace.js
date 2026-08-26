/*---Skill 工作区---*/
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

// 创建工作区文件。path 必须是工作区相对路径，后端会再次做路径和权限校验。
export const createSkillWorkspaceFile = (customSkillId, path, content = '') => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file/create`,
    method: 'post',
    data: { customSkillId, path, content },
  });
};

// 创建工作区目录。path 必须是工作区相对路径，不能是绝对服务器路径。
export const createSkillWorkspaceDirectory = (customSkillId, path) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/directory/create`,
    method: 'post',
    data: { customSkillId, path },
  });
};

// 重命名工作区条目。后端只接受当前条目的相对路径和新的文件名。
export const renameSkillWorkspaceFile = (customSkillId, path, newName) => {
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file/rename`,
    method: 'post',
    data: { customSkillId, path, newName },
  });
};

// 上传一个或多个普通文件到工作区目录。不要手动设置 Content-Type，axios
// 需要自动生成 multipart boundary；服务端不会解压归档文件。
export const uploadSkillWorkspaceFiles = (customSkillId, path, files) => {
  const formData = new FormData();
  formData.append('customSkillId', customSkillId);
  formData.append('path', path || '');
  (files || []).forEach(file => formData.append('files', file, file.name));
  return request({
    url: `${SERVICE_API}/agent/skill/workspace/file/upload`,
    method: 'post',
    data: formData,
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

/*---Skill 工作区 Git---*/
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
