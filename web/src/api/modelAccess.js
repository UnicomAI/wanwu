import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetList
export const fetchModelList = (params) => {
    return service({
        url: `${USER_API}/model/list`,
        method: "get",
        params,
    })
}

// GetsingleModel
export const getModelDetail = (params) => {
    return service({
        url: `${USER_API}/model`,
        method: "get",
        params,
    })
}

// Create
export const addModel = (data) => {
    return service({
        url: `${USER_API}/model`,
        method: "post",
        data
    })
}
// Edit
export const editModel = (data) => {
    return service({
        url: `${USER_API}/model`,
        method: "put",
        data
    })
}
// Delete
export const deleteModel = (data) => {
    return service({
        url: `${USER_API}/model`,
        method: "delete",
        data,
    })
}
// Modify status
export const changeModelStatus = (data) => {
    return service({
        url: `${USER_API}/model/status`,
        method: "put",
        data,
    })
}

//GetembeddingList
export const getEmbeddingList = (params) => {
    return service({
        url: `${USER_API}/model/select/embedding`,
        method: "get",
        params,
    })
}

//GetrerankModelList
export const getRerankList = () => {
    return service({
        url: `${USER_API}/model/select/rerank`,
        method: "get"
    })
}

// Get select model list
export const selectModelList = () => {
    return service({
        url: `${USER_API}/model/select/llm`,
        method: "get"
    })
}