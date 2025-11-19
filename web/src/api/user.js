import service from "@/utils/request";
const hasLang = true
import {USER_API} from "@/utils/requestConstants"

// Login
export const login = (data) => {
    return service({
        url: `${USER_API}/base/login`,
        method: "post",
        data,
        hasLang
    });
};

// 2FA login
// First-level verification: password
export const login2FA1 = (data) => {
    return service({
        url: `${USER_API}/base/login/email`,
        method: "post",
        data
    });
};
// Second-level verification: verification code
// Email verification code
export const login2FA2Code = (data) => {
    return service({
        url: `${USER_API}/user/login/email/code`,
        method: "post",
        data,
    });
};
// First-time login
export const login2FA2new = (data) => {
    return service({
        url: `${USER_API}/user/login`,
        method: "put",
        data
    });
}
// Returning login
export const login2FA2exist = (data) => {
    return service({
        url: `${USER_API}/user/login`,
        method: "post",
        data
    });
}

// Get graphical verification code
export const getImgVerCode = () => {
    return service({
        url: `${USER_API}/base/captcha`,
        method: "get",
        hasLang
    });
};

// Send email registration verification code
export const registerCode = (data) => {
    return service({
        url: `${USER_API}/base/register/email/code`,
        method: "post",
        data,
    });
};

// User email registration
export const register = (data) => {
    return service({
        url: `${USER_API}/base/register/email`,
        method: "post",
        data,
    });
};

// Send email verification code for password reset
export const resetCode = (data) => {
    return service({
        url: `${USER_API}/base/password/email/code`,
        method: "post",
        data,
    });
};

// ResetPassword
export const reset = (data) => {
    return service({
        url: `${USER_API}/base/password/email`,
        method: "post",
        data,
    });
};

export const getLangList = () => {
    return service({
        url: `${USER_API}/base/language/select`,
        method: "get",
    });
};


export const changeLang = (data) => {
    return service({
        url: `${USER_API}/user/language`,
        method: "put",
        data
    });
};

export const getUserDetail = (data) => {
    return service({
        url: `${USER_API}/user/info`,
        method: "get",
        params:data,
    });
};

export const getPermission = (data) => {
    return service({
        url: `${USER_API}/user/permission`,
        method: "get",
        params:data
    });
};

export const restUserPassword= (data) => {
    return service({
        url: `${USER_API}/user/admin/password`,
        method: "put",
        data,
    });
};
export const restOwnPassword= (data) => {
    return service({
        url: `${USER_API}/user/password`,
        method: "put",
        data,
    });
};

export const restAvatar= (data, config) => {
    return service({
        url: `${USER_API}/user/avatar`,
        method: "put",
        data,
        config
    });
};

export const docDownload = () => {
    return service({
        url: `${USER_API}/doc_center`,
        method: "get",
    });
};

// Shared avatar upload
export const uploadAvatar = (data, config) => {
    return service({
        url: `${USER_API}/avatar`,
        method: "post",
        data,
        config
    })
}

// Platform information
export const getCommonInfo= () => {
    return service({
        url: `${USER_API}/base/custom`,
        method: "get",
        hasLang
    });
}