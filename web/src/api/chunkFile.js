import request from "@/utils/request";
import {SERVICE_API} from "@/utils/requestConstants"
export const uploadChunks = (data,config) => {// Chunk upload
    return request({
        url: `${SERVICE_API}/file/upload`,
        method: "post",
        headers: {"Content-Type": "application/x-www-form-urlencoded"},
        data,
        cancelToken: config,
    });
}
export const checkChunks = (data) => {// Check chunk status
    return request({
        url: `${SERVICE_API}/file/check`,
        method: "get",
        params:data,
    });
}
export const mergeChunks = (data) => {// Merge chunks
    return request({
        url: `${SERVICE_API}/file/merge`,
        method: "post",
        data
    });
}
export const clearChunks = (data) => {// Clear chunks
    return request({
        url: `${SERVICE_API}/file/clean`,
        method: "post",
        data
    });
}
export const delfile = (data) => {// Delete uploaded file
    return request({
        url: `${SERVICE_API}/file/delete`,
        method: "delete",
        data
    });
}
export const continueChunks = (data) => {// Resume breakpoint uploads and fetch uploaded chunks
    return request({
        url: `${SERVICE_API}/file/check/chunk/list`,
        method: "get",
        params:data
    });
}

