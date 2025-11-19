import service from "@/utils/request";
import {USER_API} from "@/utils/requestConstants"

// Query knowledge base keyword list
export const getKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'get',
        params: data
    })
};

// Add knowledge base keywords
export const addKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'post',
        data
    })
};
// Edit knowledge base keywords
export const editKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'put',
        data
    })
};
// Delete knowledge base keywords
export const delKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'delete',
        data
    })
};
// Knowledge base keyword details
export const keyWordDetail = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords/detail`,
        method: 'get',
        params: data
    })
};