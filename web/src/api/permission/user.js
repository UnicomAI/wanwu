import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetUserList
export const fetchUserList = (params) => {
    return service({
        url: `${USER_API}/user/list`,
        method: "get",
        params,
    })
}

// GetRoleListUser
export const fetchRoleList = () => {
    return service({
        url: `${USER_API}/role/select`,
        method: "get",
    })
}
// CreateUser
export const createUser = (data) => {
    return service({
        url: `${USER_API}/user`,
        method: "post",
        data,
    })
}
// EditUser
export const editUser = (data) => {
    return service({
        url: `${USER_API}/user`,
        method: "put",
        data,
    })
}
// DeleteUser
export const deleteUser = (data) => {
    return service({
        url: `${USER_API}/user`,
        method: "delete",
        data,
    })
}
// ModifyUserStatus
export const changeUserStatus = (data) => {
    return service({
        url: `${USER_API}/user/status`,
        method: "put",
        data,
    })
}
// Get邀PleaseUser when  of UserList
export const fetchInviteUser = (params) => {
    return service({
        url: `${USER_API}/org/other/select`,
        method: "get",
        params,
    })
}
// 邀PleaseUser
export const inviteUser = (data) => {
    return service({
        url: `${USER_API}/org/user`,
        method: "post",
        data,
    })
}
