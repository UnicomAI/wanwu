import request from "@/utils/request";
import {MODEL_API, DATACENTER_API} from "@/utils/requestConstants";

/*----元景------*/
//对话List
export const getConversationList = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/list`,
        method: 'get',
        params: data
    })
};
//Create对话
export const createConversation = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/create`,
        method: 'post',
        data
    })
};
//Delete对话
export const deleteConversation = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/delete`,
        method: 'delete',
        data
    })
};
//对话Details
export const getConversationDetail = (data)=>{
    return request({
        url: `${MODEL_API}/chatllm/conversation/detail`,
        method: 'get',
        params: data
    })
};
export const addAction = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/action/create`,
        method: 'post',
        data
    })
};
export const updateAction = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/action/update`,
        method: 'put',
        data
    })
};
export const deleteAction = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/action/delete`,
        method: 'delete',
        data
    })
};
export const getActionDetail = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/action/info`,
        method: 'get',
        params: data
    })
};
export const deleteConversationHistory = (data)=>{
    return request({
        url: `${MODEL_API}/assistant/conversation/detail/delete`,
        method: 'delete',
        data
    })
};
//GetModelList
export const getModelList= (data) => {
    return request({
        url: `${DATACENTER_API}/infer/publish/model/select`,
        method: "get",
        params:data
    });
}

//AI自动Generate原生App
export const autoCreate= (data) => {
    return request({
        url: `${MODEL_API}/assistant/auto/create`,
        method: "post",
        data
    });
}

