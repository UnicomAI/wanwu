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

// 新建模板的打底配置：媒体类型没有内置（默认）模板可克隆
export const DEFAULT_TEMPLATE_CONFIG = {
  docSegment: {
    segmentMethod: '0',
    segmentType: '0',
    splitter: ['\n\n'],
    maxSplitter: 1024,
    overlap: 0.2,
    subSplitter: ['\n'],
    subMaxSplitter: 200,
  },
  docAnalyzer: ['text'],
  docPreprocess: ['replaceSymbols'],
  parserModelId: '',
  asrModelId: '',
  multimodalModelId: '',
};

// 文档类型到解析方式控件的映射，doc 走文字提取/OCR，其余走模型下拉
export const MEDIA_TYPE = {
  video: 'video',
  audio: 'audio',
  image: 'image',
};

export function getMediaType(docType) {
  return MEDIA_TYPE[docType] || 'doc';
}

// 上传文件按后缀归到解析模板的文档类型，zip/tar.gz 等归不到就返回 ''
export function getDocTypeByFileName(fileName) {
  const ext = `.${String(fileName).split('.').pop().toLowerCase()}`;
  const hit = DOC_TYPE_LIST.find(item => item.extList.includes(ext));
  return hit ? hit.docType : '';
}

// 接口用数组表达绑定，页面内部用 {docType: templateId} 取值更方便
export function bindListToMap(list) {
  return (Array.isArray(list) ? list : []).reduce((acc, item) => {
    acc[item.docType] = item.templateId;
    return acc;
  }, {});
}

// templateId 为空要照发：后端据此把该类型标成显式无模板，不再回落内置
export function bindMapToList(map) {
  return Object.entries(map || {}).map(([docType, templateId]) => ({
    docType,
    templateId: templateId || '',
  }));
}
