import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetDocument中心 md Content
export const getMarkdown = (params) => {
    return service({
        url: `${USER_API}/doc_center/markdown`,
        method: 'get',
        params
    });
};

// GetDocument中心目录
export const getDocMenu = () => {
    return service({
        url: `${USER_API}/doc_center/menu`,
        method: 'get',
    });
};

// GetDocumentSearchContent
export const getDocSearchContent = (params) => {
    return service({
        url: `${USER_API}/doc_center/search`,
        method: 'get',
        params
    });
};