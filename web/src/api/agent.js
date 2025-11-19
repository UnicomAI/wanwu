import service from "@/utils/request"
import {USER_API, OPENURL_API} from "@/utils/requestConstants"
export const createAgent = (data)=>{
    return service({
        url: `${USER_API}/assistant`,
        method: 'post',
        data
    })
}

export const updateAgent = (data)=>{
    return service({
        url: `${USER_API}/assistant`,
        method: 'put',
        data
    })
}
export const delAgent = (data)=>{
    return service({
        url: `${USER_API}/assistant`,
        method: 'delete',
        data
    })
}
export const getAgentInfo = (params)=>{
    return service({
        url: `${USER_API}/assistant`,
        method: 'get',
        params
    })
}
export const getAgentDetail = (params)=>{
    return service({
        url: `${USER_API}/assistant/draft`,
        method: 'get',
        params
    })
}
export const putAgentInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/config`,
        method: 'put',
        data
    })
}
export const createConversation = (data)=>{
    return service({
        url: `${USER_API}/assistant/conversation`,
        method: 'post',
        data
    })
}
export const delConversation = (data)=>{
    return service({
        url: `${USER_API}/assistant/conversation`,
        method: 'delete',
        data
    })
}
export const getConversationHistory = (params)=>{
    return service({
        url: `${USER_API}/assistant/conversation/detail`,
        method: 'get',
        params
    })
}
export const getConversationlist = (params)=>{
    return service({
        url: `${USER_API}/assistant/conversation/list`,
        method: 'get',
        params
    })
}
export const getActionInfo = (params)=>{
    return service({
        url: `${USER_API}/assistant/action`,
        method: 'get',
        params
    })
}
export const editActionInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/action`,
        method: 'put',
        data
    })
}
export const addActionInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/action`,
        method: 'post',
        data
    })
}
export const delActionInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/action`,
        method: 'delete',
        data
    })
}
export const enableAction = (data)=>{
    return service({
        url: `${USER_API}/assistant/action/enable`,
        method: 'put',
        data
    })
}
export const addWorkFlowInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/workflow`,
        method: 'post',
        data
    })
}
export const delWorkFlowInfo = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/workflow`,
        method: 'delete',
        data
    })
}
export const enableWorkFlow = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/workflow/switch`,
        method: 'put',
        data
    })
}
export const agentStream = (data)=>{
    return service({
        url: `${USER_API}/assistant/stream`,
        method: 'post',
        data
    })
}
export const agentTestStream = (data)=>{
    return service({
        url: `${USER_API}/assistant/test/stream`,
        method: 'post',
        data
    })
}
export const getAgentList = (params)=>{
    return service({
        url: `${USER_API}/assistant/list`,
        method: 'get',
        params
    })
}

//Delete mcp Tool
export const deleteMcp = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/mcp`,
        method: 'delete',
        data
    })
}
//Add mcp Tool
export const addMcp = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/mcp`,
        method: 'post',
        data
    })
}
//Enable/Disable mcp Tool
export const enableMcp = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/mcp/switch`,
        method: 'put',
        data
    })
}

// Delete Custom/Built-in Tool
export const delCustomBuiltIn = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool`,
        method: 'delete',
        data
    })
}
// Add Custom/Built-in Tool
export const addCustomBuiltIn = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool`,
        method: 'post',
        data
    })
}
// Enable or disable custom and built-in tools
export const switchCustomBuiltIn = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/switch`,
        method: 'put',
        data
    })
}
//ToolList
export const toolList = (data)=>{
    return service({
        url: `${USER_API}/tool/select`,
        method: 'get',
        params:data
    })
}
// Action list under a tool
export const toolActionList = (data)=>{
    return service({
        url: `${USER_API}/tool/action/list`,
        method: 'get',
        params:data
    })
}
// Action details for a built-in tool
export const toolActionDetail = (data)=>{
    return service({
        url: `${USER_API}/tool/action/detail`,
        method: 'get',
        params:data
    })
}
//mcpToolList
export const mcptoolList = (data)=>{
    return service({
        url: `${USER_API}/mcp/select`,
        method: 'get',
        params:data
    })
}
// Action list under an MCP tool
export const mcpActionList = (data)=>{
    return service({
        url: `${USER_API}/mcp/action/list`,
        method: 'get',
        params:data
    })
}
    
//Editurl
export const editOpenurl = (data)=>{
    return service({
        url: `${USER_API}/appspace/app/openurl`,
        method: 'put',
        data
    })
}
//Createurl
export const createOpenurl = (data)=>{
    return service({
        url: `${USER_API}/appspace/app/openurl`,
        method: 'post',
        data
    })
}
//DeleteAppurl
export const delOpenurl = (data)=>{
    return service({
        url: `${USER_API}/appspace/app/openurl`,
        method: 'delete',
        data
    })
}
//GetAppurlList
export const getOpenurl = (data)=>{
    return service({
        url: `${USER_API}/appspace/app/openurl/list`,
        method: 'get',
        params:data
    })
}
// Enable or disable app URL status
export const switchOpenurl = (data)=>{
    return service({
        url: `${USER_API}/appspace/app/openurl/status`,
        method: 'put',
        data
    })
}


//GetAgentopenurlInformation
export const getOpenurlInfo = (suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}`,
        method: 'get',
        ...config,
        isOpenUrl:true
    })
}
//AgentopenurlCreateAgentconversation
export const openurlConversation = (data,suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}/conversation`,
        method: 'post',
        data,
        ...config,
        isOpenUrl:true
    })
}
//DeleteAgentopenurlCreateAgentconversation
export const delOpenurlConversation = (data,suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}/conversation`,
        method: 'delete',
        data,
        ...config,
        isOpenUrl:true
    })
}
//AgentopenurlDetailshistoryList
export const OpenurlConverHistory = (data,suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}/conversation/detail`,
        method: 'get',
        params:data,
        ...config,
        isOpenUrl:true
    })
}
//AgentopenurlconversationList
export const OpenurlConverList = (suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}/conversation/list`,
        method: 'get',
        ...config,
        isOpenUrl:true
    })
}
// Agent open URL streaming conversation
export const OpenurlStream = (data,suffix,config={})=>{
    return service({
        url: `${OPENURL_API}/agent/${suffix}/stream`,
        method: 'post',
        data,
        ...config,
        isOpenUrl:true
    })
}
// Update Bocha rerank model
export const updateRerank = (data)=>{
    return service({
        url: `${USER_API}/assistant/tool/config`,
        method: 'put',
        data
    })
}