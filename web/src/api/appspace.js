import request from "@/utils/request";
import {USER_API} from "@/utils/requestConstants"

// Generateapikey
export const createApiKey = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/key`,
        method: 'post',
        data
    })
};
// Deleteapikey
export const delApiKey = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/key`,
        method: 'delete',
        data
    })
};
// GetapikeyList
export const getApiKeyList = (params)=>{
    return request({
        url: `${USER_API}/appspace/app/key/list`,
        method: 'get',
        params
    })
};
// Getapikey根Address
export const getApiKeyRoot = (params)=>{
    return request({
        url: `${USER_API}/appspace/app/url`,
        method: 'get',
        params
    })
};

// GetAgent/Text Q&A/WorkflowList
export const getAppSpaceList = (params)=>{
    return request({
        url: `${USER_API}/appspace/app/list`,
        method: 'get',
        params
    })
}

//Publishapp
export const appPublish = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/publish`,
        method: 'post',
        data
    })
};

// CancelPublishapp
export const appCancelPublish = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/publish`,
        method: 'delete',
        data
    })
};

//统一Delete工作室AppInterface
export const deleteApp = (data)=>{
    return request({
        url: `${USER_API}/appspace/app`,
        method: 'delete',
        data
    })
};

//Agent Template
export const agnetTemplateList = (params)=>{
    return request({
        url: `${USER_API}/assistant/template/list`,
        method: 'get',
        params
    })
};
//CopyAgent
export const copyAgnetTemplate = (data)=>{
    return request({
        url: `${USER_API}/assistant/template`,
        method: 'post',
        data
    })
};
//Agent TemplateDetails
export const agnetTemplateDetail = (params)=>{
    return request({
        url: `${USER_API}/assistant/template`,
        method: 'get',
        params
    })
};
//CopyText Q&AApp
export const copyTextQues = (data)=>{
    return request({
        url: `${USER_API}/appspace/rag/copy`,
        method: 'post',
        data
    })
};
//CopyAgentApp
export const copyAgentApp = (data)=>{
    return request({
        url: `${USER_API}/assistant/copy`,
        method: 'post',
        data
    })
};

