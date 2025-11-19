import request from "@/utils/request";
import {MODEL_API, SERVICE_API, USER_API, WORKFLOW_API} from "@/utils/requestConstants"

export const getWorkFlowParams = (params) => {
    return request({
        url: `${WORKFLOW_API}/workflow/parameter`,
        method: "get",
        params,
    });
};
export const useWorkFlow = (data)=>{
    return request({
        url: `${WORKFLOW_API}/api/workflow/use`,
        method: 'post',
        data
    })
};
//App广场WorkflowList
export const getExplorationFlowList = (params)=>{
    return request({
        url: `${USER_API}/exploration/app/list`,
        method: 'get',
        params
    })
};
export const createWorkFlow = (data)=>{
    return request({
        url: `${USER_API}/appspace/workflow`, //`${WORKFLOW_API}/workflow/create`,
        method: 'post',
        data
    })
};
export const copyExample = (data)=>{
    return request({
        url: `${WORKFLOW_API}/workflow/example_clone`,
        method: 'post',
        data
    })
};
export const publishWorkFlow = (data)=>{
    return request({
        url: `${WORKFLOW_API}/plugin/api/publish`,
        method: 'post',
        data
    })
};
//Copy
export const copyWorkFlow = (data)=>{
    return request({
        url: `${USER_API}/appspace/workflow/copy`, //`${WORKFLOW_API}/workflow/clone`,
        method: 'post',
        data
    })
};
//chakan
export const readWorkFlow = (data)=>{
    return request({
        url: `${WORKFLOW_API}/workflow/openapi_schema`,
        method: 'get',
        params:data
    })
};
export const externalUpload = (data, config) => {
    return request({
        // url: "/proxyupload/upload",
        url: `${SERVICE_API}/proxy/file/upload`,
        method: "post",
        data,
        config,
        isHandleRes: false
    });
};
export const getList = (data)=>{
    return request({
        url: `${MODEL_API}/mcp/select`,
        method: 'get',
        params: data
    })
};

// WorkflowImageUpload
export const uploadFile = (data) => {
    return request({
        url: `/api/bot/upload_file`,
        method: "post",
        data
    })
}

// ImportWorkflow
export const importWorkflow = (data, config) => {
    return request({
        url: `${USER_API}/appspace/workflow/import`,
        method: 'post',
        data,
        config
    });
};

// ExportWorkflow
export const exportWorkflow = (params) => {
    return request({
        url: `${USER_API}/appspace/workflow/export`,
        method: "get",
        params,
        responseType: 'blob'
    });
};