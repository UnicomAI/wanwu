import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// Get document markdown content
export const getMarkdown = (params) => {
    return service({
        url: `${USER_API}/doc_center/markdown`,
        method: 'get',
        params
    });
};

// Get document catalog
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