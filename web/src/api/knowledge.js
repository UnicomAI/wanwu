import service from "@/utils/request";
import {KNOWLEDGE_API} from "@/utils/requestConstants";
//AddDocument知识分Class
export const createDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doccategory`,
        method: "post",
        data: data,
    });
};

//ModifyDocument知识分Class
export const editDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doccategory`,
        method: "put",
        data: data,
    });
};

//DeleteDocument知识分Class
export const removeDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doccategory`,
        method: "delete",
        data: data,
    });
};

//GetDocumentList
export const getList = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/list`,
        method: "post",
        data: data,
    });
};
//EditDocument
export const modifyDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc`,
        method: "put",
        data: data,
    });
};
//DeleteDocument
export const deleteDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc`,
        method: "delete",
        data: data,
    });
};
//UploadDocument
export const importDoc = (data,source) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/import`,
        method: "post",
        cancleToken:source,
        data: data,
        headers: {"Content-Type": "multipart/form-data"}
    });
};
//SaveUploadDocument
export const saveImportDoc = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/save`,
        method: "put",
        data: data,
    });
};
//GetDocumentDownloadLink
export const getDocLink = (id) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/download_url?id=${id}`,
        method: "get"
    });
};
//命中Test
export const test = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/chunk/evaluate`,
        method: "post",
        data: data
    });
}
//UploadFileDeleteInvalidData
export const deleteInvalid = (data) => {
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/invalid`,
        method: "delete",
        data: data
    });
}
//从urlUpload
export const setUploadURL = (data)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/importUrl`,
        method: 'post',
        data
    })
};
// Parseurl
export const analyzeURL = (data)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/analysisUrl`,
        method: 'post',
        data
    })
};

// View分段ResultList
export const getContentList = (data,config)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/fileSplit`,
        method: 'post',
        data: data,
        // config
    })
};

// View分段ResultList
export const setStatus = (data,config)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/updateFileStatus`,
        method: 'post',
        data: data,
        // config
    })
};

// urlUploadBatch
export const batchurl = (data,source)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/analysisBatchUrl`,
        method: 'post',
        data: data,
        cancelToken: source,
        headers: {"Content-Type": "multipart/form-data"}
    })
};
export const batchUrlTaskStatus = (data)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/batchUrlTaskStatus`,
        method: 'get',
        params: data
    })
};
export const importBatchUrl = (data)=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/importBatchUrl`,
        method: 'get',
        params: data
    })
};
export const BatchUrlDemo = ()=>{
    return service({
        url: `${KNOWLEDGE_API}/ux/doc/downloadDemoFile`,
        method: 'get'
    })
};


//new GetKnowledge BaseList
import {USER_API} from "@/utils/requestConstants"
export const getKnowledgeList = (data)=>{
    return service({
        url: `${USER_API}/knowledge/select`,
        method: 'post',
        data
    })
};
// export const getKnowledgeItem = (params)=>{
//     return service({
//         url: `${USER_API}/knowledge`,
//         method: 'get',
//         params
//     })
// };
export const delKnowledgeItem = (data)=>{
    return service({
        url: `${USER_API}/knowledge`,
        method: 'delete',
        data
    })
};
export const createKnowledgeItem = (data)=>{
    return service({
        url: `${USER_API}/knowledge`,
        method: 'post',
        data
    })
};
export const editKnowledgeItem = (data)=>{
    return service({
        url: `${USER_API}/knowledge`,
        method: 'put',
        data
    })
};
export const getDocList = (params)=>{
    return service({
        url: `${USER_API}/knowledge/doc/list`,
        method: 'get',
        params
    })
};
export const delDocItem = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc`,
        method: 'delete',
        data
    })
};
// UploadFileTipInterface
export const uploadFileTips = (params)=>{
    return service({
        url: `${USER_API}/knowledge/doc/import/tip`,
        method: 'get',
        params
    })
};
export const getSectionList = (params)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/list`,
        method: 'get',
        params
    })
};
//UpdateDocument切片Tag
export const sectionLabels = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/labels`,
        method: 'post',
        data
    })
};
export const setSectionStatus = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/status/update`,
        method: 'post',
        data
    })
};

export const setAnalysis = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/url/analysis`,
        method: 'post',
        data
    })
};

export const docImport = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/import`,
        method: 'post',
        data
    })
};

//DeleteKnowledge BaseTag
export const delTag = (data)=>{
    return service({
        url: `${USER_API}/knowledge/tag`,
        method: 'delete',
        data
    })
};
//QueryKnowledge BaseTagList
export const tagList = (params)=>{
    return service({
        url: `${USER_API}/knowledge/tag`,
        method: 'get',
        params
    })
};
//CreateKnowledge BaseTag
export const createTag = (data)=>{
    return service({
        url: `${USER_API}/knowledge/tag`,
        method: 'post',
        data
    })
};
//ModifyKnowledge BaseTag
export const editTag = (data)=>{
    return service({
        url: `${USER_API}/knowledge/tag`,
        method: 'put',
        data
    })
};
//BindModifyKnowledge BaseTag
export const bindTag = (data)=>{
    return service({
        url: `${USER_API}/knowledge/tag/bind`,
        method: 'post',
        data
    })
};

//QueryTagBindKnowledge BaseCount
export const bindTagCount = (params)=>{
    return service({
        url: `${USER_API}/knowledge/tag/bind/count`,
        method: 'get',
        params
    })
};

//命中TestInterface
export const hitTest = (data)=>{
    return service({
        url: `${USER_API}/knowledge/hit`,
        method: 'post',
        data
    })
};
export const ocrSelectList = ()=>{
    return service({
        url: `${USER_API}/model/select/ocr`,
        method: 'get',
    })
};
export const updateDocMeta = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/meta`,
        method: 'post',
        data
    })
};
export const delSplitter = (data)=>{
    return service({
        url: `${USER_API}/knowledge/splitter`,
        method: 'delete',
        data
    })
};
export const getSplitter = (data)=>{
    return service({
        url: `${USER_API}/knowledge/splitter`,
        method: 'get',
        params:data
    })
};
export const createSplitter = (data)=>{
    return service({
        url: `${USER_API}/knowledge/splitter`,
        method: 'post',
        data
    })
};
export const editSplitter = (data)=>{
    return service({
        url: `${USER_API}/knowledge/splitter`,
        method: 'put',
        data
    })
};
export const createSegment = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/create`,
        method: 'post',
        data
    })
};
export const createBatchSegment = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/batch/create`,
        method: 'post',
        data
    })
};
export const delSegment = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/delete`,
        method: 'delete',
        data
    })
};
export const editSegment = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/update`,
        method: 'post',
        data
    })
};
export const metaSelect = (params)=>{
    return service({
        url: `${USER_API}/knowledge/meta/select`,
        method: 'get',
        params
    })
};
export const parserSelect = ()=>{
    return service({
        url: `${USER_API}/model/select/pdf-parser`,
        method: 'get'
    })
};
export const getSegmentChild = (params)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/child/list`,
        method: 'get',
        params
    })
};

export const createSegmentChild = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/child/create`,
        method: 'post',
        data
    })
};
export const delSegmentChild = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/child/delete`,
        method: 'delete',
        data
    })
};
export const updateSegmentChild = (data)=>{
    return service({
        url: `${USER_API}/knowledge/doc/segment/child/update`,
        method: 'post',
        data
    })
};
// GetKnowledge BaseOrganizationList
export const getOrgList = (data)=>{
    return service({
        url: `${USER_API}/knowledge/org`,
        method: 'get',
        params:data
    })
};
// GetKnowledge BaseOrganizationList
export const getOrgUser = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user/no/permit`,
        method: 'get',
        params:data
    })
};
// GetKnowledge BaseUserPermissionList
export const getUserPower = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user`,
        method: 'get',
        params:data
    })
};
// AddKnowledge BaseUserPermission
export const addUserPower = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user/add`,
        method: 'post',
        data
    })
};
// 转让Knowledge Base管理Permission
export const transferUserPower = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user/admin/transfer`,
        method: 'post',
        data
    })
};
// ModifyKnowledge BaseUserPermission
export const editUserPower = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user/edit`,
        method: 'post',
        data
    })
};
// DeleteKnowledge BaseUserPermission
export const delUserPower = (data)=>{
    return service({
        url: `${USER_API}/knowledge/user/delete`,
        method: 'delete',
        data
    })
};
//UpdateDocument元Data
export const updateMetaData = (data)=>{
    return service({
        url: `${USER_API}/knowledge/meta/value/update`,
        method: 'post',
        data
    })
};

//GetDocument元DataList
export const getDocMetaList = (data)=>{
    return service({
        url: `${USER_API}/knowledge/meta/value/list`,
        method: 'post',
        data
    })
};

//Get知识图谱Details
export const getGraphDetail = (data)=>{
    return service({
        url: `${USER_API}/knowledge/graph`,
        method: 'get',
        params:data
    })
};

//单条Add社Area报告
export const createCommunityReport = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/add`,
        method: 'post',
        data
    })
};
//BatchAdd社Area报告
export const createBatchCommunityReport = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/batch/add`,
        method: 'post',
        data
    })
};
//Delete社Area报告
export const delCommunityReport = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/delete`,
        method: 'delete',
        data
    })
};
//Generate社Area报告
export const generateCommunityReport = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/generate`,
        method: 'post',
        data
    })
};
//Get社Area报告
export const getCommunityReportList = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/list`,
        method: 'get',
        params:data
    })
};
//Edit社Area报告
export const editCommunityReportList = (data)=>{
    return service({
        url: `${USER_API}/knowledge/report/update`,
        method: 'post',
        data
    })
};