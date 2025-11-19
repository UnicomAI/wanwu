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

// 2FALogin
// 第一级Verify：Password
export const login2FA1 = (data) => {
    return service({
        url: `${USER_API}/base/login/email`,
        method: "post",
        data
    });
};
// 第二级Verify：Verification Code
// EmailVerification Code
export const login2FA2Code = (data) => {
    return service({
        url: `${USER_API}/user/login/email/code`,
        method: "post",
        data,
    });
};
// 首次Login
export const login2FA2new = (data) => {
    return service({
        url: `${USER_API}/user/login`,
        method: "put",
        data
    });
}
// Non首次Login
export const login2FA2exist = (data) => {
    return service({
        url: `${USER_API}/user/login`,
        method: "post",
        data
    });
}

// Get图形Verification Code
export const getImgVerCode = () => {
    return service({
        url: `${USER_API}/base/captcha`,
        method: "get",
        hasLang
    });
};

// EmailRegisterVerification CodeSend
export const registerCode = (data) => {
    return service({
        url: `${USER_API}/base/register/email/code`,
        method: "post",
        data,
    });
};

// UserEmailRegister
export const register = (data) => {
    return service({
        url: `${USER_API}/base/register/email`,
        method: "post",
        data,
    });
};

// ResetPasswordEmailVerification CodeSend
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

// 公用Upload avatar
export const uploadAvatar = (data, config) => {
    return service({
        url: `${USER_API}/avatar`,
        method: "post",
        data,
        config
    })
}

// 平台Information
export const getCommonInfo= () => {
    return service({
        url: `${USER_API}/base/custom`,
        method: "get",
        hasLang
    });
}