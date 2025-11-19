import router from '@/router/index'
import { menuList } from "@/views/layout/menu"
import { checkPerm, PERMS } from "@/router/permission"
import { i18n } from "@/lang"
import { basePath } from "@/utils/config"

export function guid() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        var r = Math.random() * 16 | 0,
            v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

export const getXClientId = () => localStorage.getItem('xClientId')

// Used forLogin切OrganizationEtc找到HasPermission of 第一个MenuPath (除用Model：用Model为打开 of 新Page)
export const fetchPermFirPath = (list = menuList) => {
    if (!list.length) return ''

    let path = ''
    for (let i in list) {
        const item = list[i]

        if (item.path && checkPerm(item.perm)) {
            if (item.children && item.children.length) {
                path = fetchPermFirPath(item.children).path
                break
            } else {
                path = item.path || ''
                break
            }
        }
    }

    // IfHasPermission，跳转左侧Menu第一个HasPermission of Page；NoThen跳转 /404
    return {path: path || '/404'}
}

// 找到HasPermission of 第一个Menu of  index
export const fetchCurrentPathIndex = (path, list) => {
    let index = ''
    const findIndex = (list) => {
        for (let i in list) {
            let item = list[i]
            const formatPath = (url) => url + '/'
            if (item.path && formatPath(path).includes(formatPath(item.path))) {
                index = item.index
            } else {
                if (item.children && item.children.length) {
                    findIndex(item.children)
                }
            }
        }
        return index
    }
    return findIndex(list)
}

export const jumpPermUrl = () => {
    const {path} = fetchPermFirPath()

    router.push({path: path || '/404'})
}

export const jumpOAuth = (params) => {
    router.push({
        path: "/oauth",
        query: params
    });
}

export const redirectUrl = () => {
    // 跳到HasPermission of 第一个Page
    jumpPermUrl()
}

export const replaceIcon = (logoPath) => {
    let link = document.querySelector("link[rel*='icon']") || document.createElement("link")
    link.type = "image/x-icon"
    link.rel = "shortcut icon"
    link.href = logoPath ? basePath + '/user/api' + logoPath : basePath + '/aibase/favicon.ico'
    document.getElementsByTagName("head")[0].appendChild(link)
}

export const replaceTitle = (title) => {
    document.title = title || i18n.t('header.title')
}

export const getInitTimeRange = () => {
    const date = new Date()
    const month = date.getMonth() + 1
    const startTime =  date.getFullYear() + "-" + (month < 10 ? "0" : "") + month + "-"  + '01 00:00:00'
    const stamp= new Date().getTime() + 8 * 60 * 60 * 1000
    const endTime = new Date(stamp).toISOString().replace(/T/, ' ').replace(/\..+/, '').substring(0, 19)
    return [startTime, endTime]
}

export function convertLatexSyntax(inputText) {
    // 1. Match块级公式，将 `\[` And `\]` Replace为 `$$`，支持 `\\[` `\\]` OR单个 `\[` `\]`
    inputText = inputText.replace(/\\\[\s*([\s\S]+?)\s*\\\]/g, (_, formula) => `$$${formula}$$`);
    // 2. Match行内公式，将 `\(` And `\)` Replace为 `$`，支持 `\\(` `\\)` OR单个 `\(` `\)`
    inputText = inputText.replace(/\\\(\s*([\s\S]+?)\s*\\\)/g, (_, formula) => `$${formula}$`);
    return inputText;
}

export function isSub(data){
    return /\【([0-9]{0,2})\^\】/.test(data)
}

export function parseSub(data,index){
    return data.replace(/\【([0-9]{0,2})\^\】/g,(item)=>{
        let result = item.match(/\【([0-9]{0,2})\^\】/)[1]
        return `<sup class='citation' data-parents-index='${index}'>${result}</sup>`
    })
}

/**
 *GetURLParameter
 */
export function getQueryString(val, href) {
    const hrefNew = href || window.location.href;
    const search = hrefNew.substring(hrefNew.lastIndexOf('?') + 1, hrefNew.length);
    // 组装?
    const uri = '?' + search;
    const reg = new RegExp('' + val + '=([^&?]*)', 'ig');
    const matchArr = uri.match(reg);
    if (matchArr && matchArr.length) {
        return matchArr[0].substring(val.length + 1);
    }
    return null;
}

// YesNoYesHas效 of URL
export function isValidURL(string) {
    const res = string.match(/(https?|ftp|file|ssh):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|]/i);
    return res !== null;
}

export function isExternal(path) {
    return /^(https?:|mailto:|tel:)/.test(path);
}

export const formatTools = (tools) => {
    if (!(tools && tools.length)) return []
    const newTools = tools.map((n,i)=>{
        let params = []
        let properties = n.inputSchema.properties
        for(let key in properties){
            params.push({
                "name": key,
                "requiredBadge": n.inputSchema.required && n.inputSchema.required.includes(key) ? 'Required' : '',
                "type": properties[key].type,
                "description": properties[key].description,
            })
        }
        return {
            ...n,
            params
        }
    })
    return newTools
}

/**
 * 格式化得分，保留5位小数
 * @param {number|string} score - 得分Value
 * @returns {string} 格式化后 of 得分String
 */
export function formatScore(score) {
    // Format得分，保留5位小数
    if (typeof score !== 'number') {
        return '0.00000';
    }
    return score.toFixed(5);
}

export function avatarSrc(path){
    return basePath + '/user/api/' + path
}

// 换算单位万/亿/万亿，保留2位小数
export const formatAmount = (num, returnType = 'string', preserveRange = false) => {
    const units = i18n.t("statisticsEcharts.units");
    const isHasDecimal = num.toString().includes('.');
    let formatNum = num
    let simplifiedNum = num.toString();

    // 99999以内原样Show
    if (preserveRange && num < 100000) {
        if (returnType === 'object') {
            return {
                value: simplifiedNum,
                type: ''
            };
        } else {
            return simplifiedNum;
        }
    }

    if (isHasDecimal) {
        formatNum = Number(num.toString().slice(0, num.toString().indexOf('.')))
    }
    // GetNumber of Count级
    let unitIndex = Math.floor((String(formatNum).length - 1) / 4);

    if (unitIndex > 0) {
        const unit = units[unitIndex];

        const divisor = Math.pow(10, unitIndex * 4);
        //缩小相应倍数，并保留2位小数
        const formattedValue = (num / divisor).toFixed(2).replace(/(\d)(?=(\d{3})+(?!\d))/g, "$1,");

        if (returnType === 'object') {
            return {
                value: formattedValue,
                type: unit
            };
        } else {
            simplifiedNum = formattedValue + unit;
        }
    } else if (returnType === 'object') {
        // Count级为0 when  of ObjectFormatBack
        return {
            value: simplifiedNum,
            type: ''
        };
    }

    return simplifiedNum;
}

