import request from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

/*---WorkflowTemplate---*/
export const getWorkflowTempList = (data)=>{
    return request({
        url: `${USER_API}/workflow/template/list`,
        method: 'get',
        params: data
    })
};
export const getWorkflowTempInfo = (data)=>{
    return request({
        url: `${USER_API}/workflow/template/detail`,
        method: 'get',
        params: data
    })
};
export const getWorkflowRecommendsList = (data)=>{
    return request({
        url: `${USER_API}/workflow/template/recommend`,
        method: 'get',
        params: data
    })
};
export const downloadWorkflow = (params) => {
    return request({
        url: `${USER_API}/workflow/template/download`,
        method: "get",
        params,
        responseType: 'blob'
    });
};
export const copyWorkflowTemplate = (data)=>{
    return request({
        url: `${USER_API}/workflow/template`,
        method: 'post',
        data
    })
};

/*---Tip词Template---*/
export const getPromptTempList = (data)=>{
    return request({
        url: `${USER_API}/prompt/template/list`,
        method: 'get',
        params: data
    })
};

export const copyPromptTemplate = (data)=>{
    return request({
        url: `${USER_API}/prompt/template`,
        method: 'post',
        data
    })
};

/*---CustomTip词---*/
export const getCustomPromptList = (data)=>{
    return request({
        url: `${USER_API}/prompt/custom/list`,
        method: 'get',
        params: data
    })
};

export const createCustomPrompt = (data)=>{
    return request({
        url: `${USER_API}/prompt/custom`,
        method: 'post',
        data
    })
};

export const editCustomPrompt = (data)=>{
    return request({
        url: `${USER_API}/prompt/custom`,
        method: 'put',
        data
    })
};

export const copyCustomPrompt = (data)=>{
    return request({
        url: `${USER_API}/prompt/custom/copy`,
        method: 'post',
        data
    })
};

export const deleteCustomPrompt = (data)=>{
    return request({
        url: `${USER_API}/prompt/custom`,
        method: 'delete',
        data
    })
};