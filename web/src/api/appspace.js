import request from "@/utils/request";
import {USER_API} from "@/utils/requestConstants"

// Generate API key
export const createApiKey = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/key`,
        method: 'post',
        data
    })
};
// Delete API key
export const delApiKey = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/key`,
        method: 'delete',
        data
    })
};
// Get API key list
export const getApiKeyList = (params)=>{
    return request({
        url: `${USER_API}/appspace/app/key/list`,
        method: 'get',
        params
    })
};
// Get API key base URL
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

// Publish app
export const appPublish = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/publish`,
        method: 'post',
        data
    })
};

// Cancel app publication
export const appCancelPublish = (data)=>{
    return request({
        url: `${USER_API}/appspace/app/publish`,
        method: 'delete',
        data
    })
};

// Unified deletion interface for studio apps
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
// Copy text Q&A app
export const copyTextQues = (data)=>{
    return request({
        url: `${USER_API}/appspace/rag/copy`,
        method: 'post',
        data
    })
};
// Copy agent app
export const copyAgentApp = (data)=>{
    return request({
        url: `${USER_API}/assistant/copy`,
        method: 'post',
        data
    })
};

