import request from "@/utils/request";
import {MODEL_API, SERVICE_API} from "@/utils/requestConstants"

export const createApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/create`,
        method: 'post',
        data
    })
};
export const updateApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/update`,
        method: 'put',
        data
    })
};
export const getAppDetail = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/info`,
        method: 'get',
        params: data
    })
};
export const deleteApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/delete`,
        method: 'delete',
        data
    })
};
export const publishApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/publish`,
        method: 'post',
        data
    })
};
export const getAppDraftList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/draft_list`,
        method: 'get',
        params: data
    })
};
export const getAppMoreList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/more_list`,
        method: 'get',
        params: data
    })
};
export const getMyAppList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/list`,
        method: 'get',
        params: data
    })
};
//AvatarUpload
export const fileUpload = (data,config)=>{
    return request({
        url: `${SERVICE_API}/model/expansion/file/batch/upload`,
        method: 'post',
        data,
        config
    })
};
//知识增强FileUpload
export const knowledgeFileUpload = (data,config)=>{
    return request({
        url: `${MODEL_API}/assistant/knowledge/file/upload`,
        method: 'post',
        data,
        config
    })
};
//Query已UploadFileList
export const getKnowledgeFileList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/knowledge/file/list`,
        method: 'get',
        params: data
    })
};
export const deleteKnowledgeFile = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/knowledge/file/delete`,
        method: 'delete',
        data
    })
};
//常用App
export const getRecentApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/common/list`,
        method: 'get',
        params: data
    })
};
//Delete常用App
export const deleteRecentApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/common/delete`,
        method: 'delete',
        data
    })
};

//对话List
export const getConversationList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/list`,
        method: 'get',
        params: data
    })
};
//Create对话
export const createConversation = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/create`,
        method: 'post',
        data
    })
};
//Delete对话
export const deleteConversation = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/delete`,
        method: 'delete',
        data
    })
};
//对话Details
export const getConversationDetail = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/detail`,
        method: 'get',
        params: data
    })
};
/*----元景------*/
//对话List
export const getConversationListCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/list`,
        method: 'get',
        params: data
    })
};
//Create对话
export const createConversationCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/create`,
        method: 'post',
        data
    })
};
//Delete对话
export const deleteConversationCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/delete`,
        method: 'delete',
        data
    })
};
//对话Details
export const getConversationDetailCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/detail`,
        method: 'get',
        params: data
    })
};
//BatchFileUpload
export const batchUpload = (data,config)=>{
    return request({
        url: `${MODEL_API}/file/batch/upload`,
        method: 'post',
        data,
        config
    })
};
// app接入
export const linkAPP = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/app/publish`,
        method: 'post',
        data
    })
};

//推荐AgentList
export const recommendList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/recommend/list`,
        method: 'get',
        params:data
    })
};
//标记推荐Agent
export const recommendMark = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/recommend/update`,
        method: 'put',
        data
    })
};

//UploadFileConfirmPath
export const confirmPath = (data)=>{
    return request({
        url: `${MODEL_API}/file/confirmPath`,
        method: 'post',
        data
    })
};
