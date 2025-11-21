import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetOrganizationList
export const fetchOrgList = (params) => {
    return service({
        url: `${USER_API}/org/list`,
        method: "get",
        params,
    })
}
// GetOrganizationDetails
export const fetchOrgDetail = (params) => {
    return service({
        url: `${USER_API}/org/info`,
        method: "get",
        params,
    })
}
// CreateOrganization
export const createOrg = (data) => {
    return service({
        url: `${USER_API}/org`,
        method: "post",
        data,
    })
}
// EditOrganization
export const editOrg = (data) => {
    return service({
        url: `${USER_API}/org`,
        method: "put",
        data,
    })
}
// DeleteOrganization
export const deleteOrg = (data) => {
    return service({
        url: `${USER_API}/org`,
        method: "delete",
        data,
    })
}
// ModifyOrganizationStatus
export const changeOrgStatus = (data) => {
    return service({
        url: `${USER_API}/org/status`,
        method: "put",
        data,
    })
}

// GetNavigationOrganizationList
export const fetchOrgs = () => {
    return service({
        url: `${USER_API}/org/select`,
        method: "get",
    })
}