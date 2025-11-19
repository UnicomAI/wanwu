import request from "@/utils/request";
import {USER_API} from "@/utils/requestConstants"

//EditSensitive Word Table
export const editSensitive = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/table`,
        method: 'put',
        data
    })
};
//CreateSensitive Word Table
export const createSensitive = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/table`,
        method: 'post',
        data
    })
};
//DeleteSensitive Word Table
export const delSensitive = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/table`,
        method: 'delete',
        data
    })
};
//ViewSensitive Word TableList
export const getSensitiveList = ()=>{
    return request({
        url: `${USER_API}/safe/sensitive/table/list`,
        method: 'get',
    })
};
//EditReplySetting
export const setReply = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/table/reply`,
        method: 'put',
        data
    })
};
//GetSensitive Word Table下拉List
export const sensitiveSelect = ()=>{
    return request({
        url: `${USER_API}/safe/sensitive/table/select`,
        method: 'get',
    })
};
//DeleteSensitive Word
export const delSensitiveWord = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/word`,
        method: 'delete',
        data
    })
};
//Query词TableDataList
export const getSensitiveWord = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/word/list`,
        method: 'get',
        params:data
    })
};
//UploadSensitive Word
export const uploadSensitiveWord = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/word`,
        method: 'post',
        data
    })
};
//GetSensitive WordReplySetting
export const getReplay = (data)=>{
    return request({
        url: `${USER_API}/safe/sensitive/table`,
        method: 'get',
        params:data
    })
};