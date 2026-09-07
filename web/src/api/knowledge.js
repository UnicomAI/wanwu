import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

export const getKnowledgeList = data => {
  return service({
    url: `${USER_API}/knowledge/select`,
    method: 'post',
    data,
  });
};
// export const getKnowledgeItem = (params)=>{
//     return service({
//         url: `${USER_API}/knowledge`,
//         method: 'get',
//         params
//     })
// };
export const delKnowledgeItem = data => {
  return service({
    url: `${USER_API}/knowledge`,
    method: 'delete',
    data,
  });
};
export const createKnowledgeItem = data => {
  return service({
    url: `${USER_API}/knowledge`,
    method: 'post',
    data,
  });
};
export const editKnowledgeItem = data => {
  return service({
    url: `${USER_API}/knowledge`,
    method: 'put',
    data,
  });
};
export const getDocDetail = params => {
  return service({
    url: `${USER_API}/knowledge/doc/detail`,
    method: 'get',
    params,
  });
};
export const getDocList = data => {
  return service({
    url: `${USER_API}/knowledge/doc/list`,
    method: 'post',
    data,
  });
};
export const delDocItem = data => {
  return service({
    url: `${USER_API}/knowledge/doc`,
    method: 'delete',
    data,
  });
};
// 知识库文档导出
export const exportDoc = data => {
  return service({
    url: `${USER_API}/knowledge/doc/export`,
    method: 'post',
    data,
  });
};
// 上传文件提示接口
export const uploadFileTips = params => {
  return service({
    url: `${USER_API}/knowledge/doc/import/tip`,
    method: 'get',
    params,
  });
};
export const getSectionList = params => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/list`,
    method: 'get',
    params,
  });
};
//更新文档切片标签
export const sectionLabels = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/labels`,
    method: 'post',
    data,
  });
};
export const setSectionStatus = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/status/update`,
    method: 'post',
    data,
  });
};

export const setAnalysis = data => {
  return service({
    url: `${USER_API}/knowledge/doc/url/analysis`,
    method: 'post',
    data,
  });
};

export const docImport = data => {
  return service({
    url: `${USER_API}/knowledge/doc/import`,
    method: 'post',
    data,
  });
};

export const docReImport = data => {
  return service({
    url: `${USER_API}/knowledge/doc/reimport`,
    method: 'post',
    data,
  });
};

export const getDocLimit = params => {
  return service({
    url: `${USER_API}/knowledge/doc/upload/limit`,
    method: 'get',
    params,
  });
};

export const getDocConfig = params => {
  return service({
    url: `${USER_API}/knowledge/doc/config`,
    method: 'get',
    params,
  });
};

export const updateDocConfig = data => {
  return service({
    url: `${USER_API}/knowledge/doc/update/config`,
    method: 'post',
    data,
  });
};

//删除知识库标签
export const delTag = data => {
  return service({
    url: `${USER_API}/knowledge/tag`,
    method: 'delete',
    data,
  });
};
//查询知识库标签列表
export const tagList = params => {
  return service({
    url: `${USER_API}/knowledge/tag`,
    method: 'get',
    params,
  });
};
//创建知识库标签
export const createTag = data => {
  return service({
    url: `${USER_API}/knowledge/tag`,
    method: 'post',
    data,
  });
};
//修改知识库标签
export const editTag = data => {
  return service({
    url: `${USER_API}/knowledge/tag`,
    method: 'put',
    data,
  });
};
//绑定修改知识库标签
export const bindTag = data => {
  return service({
    url: `${USER_API}/knowledge/tag/bind`,
    method: 'post',
    data,
  });
};

//查询标签绑定知识库数量
export const bindTagCount = params => {
  return service({
    url: `${USER_API}/knowledge/tag/bind/count`,
    method: 'get',
    params,
  });
};

//命中测试接口
export const hitTest = data => {
  return service({
    url: `${USER_API}/knowledge/hit`,
    method: 'post',
    data,
  });
};
export const ocrSelectList = () => {
  return service({
    url: `${USER_API}/model/select/ocr`,
    method: 'get',
  });
};
export const updateDocMeta = data => {
  return service({
    url: `${USER_API}/knowledge/doc/meta`,
    method: 'post',
    data,
  });
};
export const delSplitter = data => {
  return service({
    url: `${USER_API}/knowledge/splitter`,
    method: 'delete',
    data,
  });
};
export const getSplitter = data => {
  return service({
    url: `${USER_API}/knowledge/splitter`,
    method: 'get',
    params: data,
  });
};
export const createSplitter = data => {
  return service({
    url: `${USER_API}/knowledge/splitter`,
    method: 'post',
    data,
  });
};
export const editSplitter = data => {
  return service({
    url: `${USER_API}/knowledge/splitter`,
    method: 'put',
    data,
  });
};
export const createSegment = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/create`,
    method: 'post',
    data,
  });
};
export const createBatchSegment = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/batch/create`,
    method: 'post',
    data,
  });
};
export const delSegment = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/delete`,
    method: 'delete',
    data,
  });
};
export const editSegment = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/update`,
    method: 'post',
    data,
  });
};
export const metaSelect = params => {
  return service({
    url: `${USER_API}/knowledge/meta/select`,
    method: 'get',
    params,
  });
};
export const parserSelect = () => {
  return service({
    url: `${USER_API}/model/select/pdf-parser`,
    method: 'get',
  });
};
export const getSegmentChild = params => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/child/list`,
    method: 'get',
    params,
  });
};

export const createSegmentChild = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/child/create`,
    method: 'post',
    data,
  });
};
export const delSegmentChild = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/child/delete`,
    method: 'delete',
    data,
  });
};
export const updateSegmentChild = data => {
  return service({
    url: `${USER_API}/knowledge/doc/segment/child/update`,
    method: 'post',
    data,
  });
};
// 获取知识库组织列表
export const getOrgList = data => {
  return service({
    url: `${USER_API}/knowledge/org`,
    method: 'get',
    params: data,
  });
};
// 获取知识库组织列表
export const getOrgUser = data => {
  return service({
    url: `${USER_API}/knowledge/user/no/permit`,
    method: 'get',
    params: data,
  });
};
// 获取知识库用户权限列表
export const getUserPower = data => {
  return service({
    url: `${USER_API}/knowledge/user`,
    method: 'get',
    params: data,
  });
};
// 新增知识库用户权限
export const addUserPower = data => {
  return service({
    url: `${USER_API}/knowledge/user/add`,
    method: 'post',
    data,
  });
};
// 转让知识库管理权限
export const transferUserPower = data => {
  return service({
    url: `${USER_API}/knowledge/user/admin/transfer`,
    method: 'post',
    data,
  });
};
// 修改知识库用户权限
export const editUserPower = data => {
  return service({
    url: `${USER_API}/knowledge/user/edit`,
    method: 'post',
    data,
  });
};
// 删除知识库用户权限
export const delUserPower = data => {
  return service({
    url: `${USER_API}/knowledge/user/delete`,
    method: 'delete',
    data,
  });
};
//更新文档元数据
export const updateMetaData = data => {
  return service({
    url: `${USER_API}/knowledge/meta/value/update`,
    method: 'post',
    data,
  });
};

//获取文档元数据列表
export const getDocMetaList = data => {
  return service({
    url: `${USER_API}/knowledge/meta/value/list`,
    method: 'post',
    data,
  });
};

//获取知识图谱详情
export const getGraphDetail = data => {
  return service({
    url: `${USER_API}/knowledge/graph`,
    method: 'get',
    params: data,
  });
};

//单条新增社区报告
export const createCommunityReport = data => {
  return service({
    url: `${USER_API}/knowledge/report/add`,
    method: 'post',
    data,
  });
};
//批量新增社区报告
export const createBatchCommunityReport = data => {
  return service({
    url: `${USER_API}/knowledge/report/batch/add`,
    method: 'post',
    data,
  });
};
//删除社区报告
export const delCommunityReport = data => {
  return service({
    url: `${USER_API}/knowledge/report/delete`,
    method: 'delete',
    data,
  });
};
//生成社区报告
export const generateCommunityReport = data => {
  return service({
    url: `${USER_API}/knowledge/report/generate`,
    method: 'post',
    data,
  });
};
//获取社区报告
export const getCommunityReportList = data => {
  return service({
    url: `${USER_API}/knowledge/report/list`,
    method: 'get',
    params: data,
  });
};
//编辑社区报告
export const editCommunityReportList = data => {
  return service({
    url: `${USER_API}/knowledge/report/update`,
    method: 'post',
    data,
  });
};

export const createExternal = data => {
  return service({
    url: `${USER_API}/knowledge/external`,
    method: 'post',
    data,
  });
};

export const editExternal = data => {
  return service({
    url: `${USER_API}/knowledge/external`,
    method: 'put',
    data,
  });
};

export const delExternal = data => {
  return service({
    url: `${USER_API}/knowledge/external`,
    method: 'delete',
    data,
  });
};

export const importExternalAPI = data => {
  return service({
    url: `${USER_API}/knowledge/external/api`,
    method: 'post',
    data,
  });
};

export const editExternalAPI = data => {
  return service({
    url: `${USER_API}/knowledge/external/api`,
    method: 'put',
    data,
  });
};

export const delExternalAPI = data => {
  return service({
    url: `${USER_API}/knowledge/external/api`,
    method: 'delete',
    data,
  });
};

export const getExternalAPIList = data => {
  return service({
    url: `${USER_API}/knowledge/external/api/select`,
    method: 'get',
    params: data,
  });
};

export const getExternalList = data => {
  return service({
    url: `${USER_API}/knowledge/external/select`,
    method: 'get',
    params: data,
  });
};

// 根据知识库名称和文档名称获取文档信息
export const getDocByName = data => {
  return service({
    url: `${USER_API}/knowledge/doc/by/name`,
    method: 'get',
    params: data,
  });
};
