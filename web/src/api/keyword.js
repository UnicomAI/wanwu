import service from "@/utils/request";
import {USER_API} from "@/utils/requestConstants"

//QueryKnowledge Base关Key词List
export const getKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'get',
        params: data
    })
};

//AddKnowledge Base关Key词List
export const addKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'post',
        data
    })
};
//EditKnowledge Base关Key词List
export const editKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'put',
        data
    })
};
//DeleteKnowledge Base关Key词List
export const delKeyWord = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords`,
        method: 'delete',
        data
    })
};
//Knowledge Base关Key词Details
export const keyWordDetail = (data)=>{
    return service({
        url: `${USER_API}/knowledge/keywords/detail`,
        method: 'get',
        params: data
    })
};