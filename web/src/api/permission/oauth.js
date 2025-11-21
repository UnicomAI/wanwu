import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetOAuthAppList
export const fetchOAuthList = (data) => {
    return service({
        url: `${USER_API}/oauth/app/list`,
        method: "get",
        params: data,
    })
}

// CreateOAuthApp
export const createOAuth = (data) => {
    return service({
        url: `${USER_API}/oauth/app`,
        method: "post",
        data: data,
    })
}

// UpdateOAuthApp
export const editOAuth = (data) => {
    return service({
        url: `${USER_API}/oauth/app`,
        method: "put",
        data: data,
    })
}

// DeleteOAuthApp
export const deleteOAuth = (data) => {
    return service({
        url: `${USER_API}/oauth/app`,
        method: "delete",
        data: data,
    })
}

// ModifyOAuthAppStatus
export const changeOAuthStatus = (data) => {
    return service({
        url: `${USER_API}/oauth/app/status`,
        method: "put",
        data: data,
    })
}

// Authorization code authentication
export const codeOAuth = (data) => {
    return service({
        url: `${USER_API}/oauth/code/authorize`,
        method: "get",
        params: data,
    })
}