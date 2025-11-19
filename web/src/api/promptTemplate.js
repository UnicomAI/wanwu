import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"
// Get custom prompt details
export const getPromptTemplateDetail = (data)=>{
    return service({
        url: `${USER_API}/prompt/custom`,
        method: 'get',
        params: data
    })
}
// Get custom prompt list
export const getPromptTemplateList= (data)=>{
    return service({
        url: `${USER_API}/prompt/custom/list`,
        method: 'get',
        params: data
    })
}

// Get built-in prompt list
export const getPromptBuiltInList= (data)=>{
    return service({
        url: `${USER_API}/prompt/template/list`,
        method: 'get',
        params: data
    })
}
// Get built-in prompt details
export const getPromptBuiltInDetail= (data)=>{
    return service({
        url: `${USER_API}/prompt/template/detail`,
        method: 'get',
        params: data
    })
}

