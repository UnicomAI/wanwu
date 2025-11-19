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
// Knowledge enhancement file upload
export const knowledgeFileUpload = (data,config)=>{
    return request({
        url: `${MODEL_API}/assistant/knowledge/file/upload`,
        method: 'post',
        data,
        config
    })
};
//QueryalreadyUploadFileList
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
// Frequently used apps
export const getRecentApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/common/list`,
        method: 'get',
        params: data
    })
};
// Delete frequently used apps
export const deleteRecentApp = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/common/delete`,
        method: 'delete',
        data
    })
};

//conversationList
export const getConversationList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/list`,
        method: 'get',
        params: data
    })
};
//Createconversation
export const createConversation = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/create`,
        method: 'post',
        data
    })
};
//Delete Conversation
export const deleteConversation = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/delete`,
        method: 'delete',
        data
    })
};
//conversationDetails
export const getConversationDetail = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/detail`,
        method: 'get',
        params: data
    })
};
/*---- Yuanjing ------*/
//conversationList
export const getConversationListCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/list`,
        method: 'get',
        params: data
    })
};
//Createconversation
export const createConversationCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/create`,
        method: 'post',
        data
    })
};
//Delete Conversation
export const deleteConversationCUBM = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/delete`,
        method: 'delete',
        data
    })
};
//conversationDetails
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
// App integration
export const linkAPP = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/app/publish`,
        method: 'post',
        data
    })
};

// Recommended agent list
export const recommendList = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/recommend/list`,
        method: 'get',
        params:data
    })
};
// Mark recommended agent
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
