import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"
//GetCustompropmptDetails
export const getPromptTemplateDetail = (data)=>{
    return service({
        url: `${USER_API}/prompt/custom`,
        method: 'get',
        params: data
    })
}
//GetCustompromptList
export const getPromptTemplateList= (data)=>{
    return service({
        url: `${USER_API}/prompt/custom/list`,
        method: 'get',
        params: data
    })
}

//Get内置promptList
export const getPromptBuiltInList= (data)=>{
    return service({
        url: `${USER_API}/prompt/template/list`,
        method: 'get',
        params: data
    })
}
//Get内置promptDetails
export const getPromptBuiltInDetail= (data)=>{
    return service({
        url: `${USER_API}/prompt/template/detail`,
        method: 'get',
        params: data
    })
}

