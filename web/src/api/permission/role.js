import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetRoleList
export const fetchRoleList = (params) => {
    return service({
        url: `${USER_API}/role/list`,
        method: "get",
        params,
    })
}
// GetRoleDetails
export const fetchRoleDetail = (params) => {
    return service({
        url: `${USER_API}/role/info`,
        method: "get",
        params,
    })
}
// CreateRole
export const createRole = (data) => {
    return service({
        url: `${USER_API}/role`,
        method: "post",
        data,
    })
}
// EditRole
export const editRole = (data) => {
    return service({
        url: `${USER_API}/role`,
        method: "put",
        data,
    })
}
// DeleteRole
export const deleteRole = (data) => {
    return service({
        url: `${USER_API}/role`,
        method: "delete",
        data,
    })
}
// ModifyRoleStatus
export const changeRoleStatus = (data) => {
    return service({
        url: `${USER_API}/role/status`,
        method: "put",
        data,
    })
}
// GetPermissionTree
export const fetchPermTree = () => {
    return service({
        url: `${USER_API}/role/template`,
        method: "get",
    })
}
