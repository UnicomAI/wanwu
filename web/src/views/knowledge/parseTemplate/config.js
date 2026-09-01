import { i18n } from '@/lang';

// 解析模板支持的文档类型，extList 用于卡片上展示后缀
export const DOC_TYPE_LIST = [
  { docType: 'pdf', icon: 'pdf', extList: ['.pdf'] },
  { docType: 'word', icon: 'docx', extList: ['.docx', '.doc'] },
  { docType: 'ppt', icon: 'pptx', extList: ['.pptx'] },
  { docType: 'excel', icon: 'xlsx', extList: ['.xlsx', '.xls'] },
  { docType: 'csv', icon: 'csv', extList: ['.csv'] },
  { docType: 'txt', icon: 'txt', extList: ['.txt'] },
  { docType: 'html', icon: 'html', extList: ['.html'] },
  { docType: 'markdown', icon: 'markdown', extList: ['.md'] },
  { docType: 'wps', icon: 'doc', extList: ['.wps'] },
  { docType: 'ofd', icon: 'ofd', extList: ['.ofd'] },
  { docType: 'video', icon: 'mp4', extList: ['.avi', '.mp4', '.mov', '.wmv'] },
  { docType: 'audio', icon: 'mp3', extList: ['.mp3', '.wav', '.aac', '.m4a'] },
  { docType: 'image', icon: 'png', extList: ['.png', '.jpg', '.jpeg'] },
].map(item => ({
  ...item,
  name: i18n.t(`knowledgeManage.parseTemplate.docTypeName.${item.docType}`),
}));

// 文档类型到解析方式控件的映射，doc 走文字提取/OCR，其余走模型下拉
export const MEDIA_TYPE = {
  video: 'video',
  audio: 'audio',
  image: 'image',
};

export function getMediaType(docType) {
  return MEDIA_TYPE[docType] || 'doc';
}

// 接口用数组表达绑定，页面内部用 {docType: templateId} 取值更方便
export function bindListToMap(list) {
  return (Array.isArray(list) ? list : []).reduce((acc, item) => {
    acc[item.docType] = item.templateId;
    return acc;
  }, {});
}

export function bindMapToList(map) {
  return Object.entries(map || {})
    .filter(([, templateId]) => templateId)
    .map(([docType, templateId]) => ({ docType, templateId }));
}
