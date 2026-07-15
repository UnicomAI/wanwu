import router from '@/router/index';
import { menuList } from '@/views/layout/menu';
import { checkPerm } from '@/router/permission';
import { i18n } from '@/lang';
import { Message } from 'element-ui';
import { isSafeImageUrl, escapeHtml as sanitizeEscapeHtml } from './sanitize';
import { basePath } from '@/utils/config';
import { store } from '@/store';

export function guid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replaceAll(
    /[xy]/g,
    function (c) {
      let r = Math.trunc(Math.random() * 16),
        v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    },
  );
}

export const getXClientId = () => localStorage.getItem('xClientId');

// 用于登录切组织等找到有权限的第一个菜单路径
export const fetchPermFirPath = (list = menuList) => {
  if (!list.length) return { path: '/404' };

  let path = '';
  for (let i in list) {
    const item = list[i];

    if (checkPerm(item.perm)) {
      if (item.children?.length) {
        path = fetchPermFirPath(item.children).path;
        break;
      } else {
        path = item.path || '';
        break;
      }
    }
  }

  // 若有权限，跳转左侧菜单第一个有权限的页面；否则跳转 /404
  return { path: path || '/404' };
};

// 找到有权限的第一个菜单的 index
export const fetchCurrentPathIndex = (path, list) => {
  let index = '';
  const findIndex = list => {
    for (let i in list) {
      let item = list[i];
      const formatPath = url => {
        // 对于 文本问答/工作流/智能体 前面带了 /appSpace 特殊路由的处理
        if (url.includes('/appSpace/')) {
          return url.slice(9) + '/';
        }
        return url + '/';
      };
      if (item.path && formatPath(path).includes(formatPath(item.path))) {
        index = item.index;
      } else if (item.children?.length) {
        findIndex(item.children);
      }
    }
    return index;
  };
  return findIndex(list);
};

export const jumpPermUrl = () => {
  const { path } = fetchPermFirPath();

  router.push({ path: path || '/404' });
};

export const jumpOAuth = params => {
  router.push({
    path: '/oauth',
    query: params,
  });
};

export const redirectUrl = () => {
  // 跳到有权限的第一个页面
  jumpPermUrl();
};

export const redirectUserInfoPage = (
  isUpdatePassword,
  callback,
  isRedirectUrl,
) => {
  if (isUpdatePassword !== undefined && !isUpdatePassword) {
    router.push('/userInfo?showPwd=1');
    callback?.();
  } else if (isRedirectUrl) jumpPermUrl();
};

export const replaceIcon = logoPath => {
  let link =
    document.querySelector("link[rel*='icon']") ||
    document.createElement('link');
  link.type = 'image/x-icon';
  link.rel = 'shortcut icon';
  link.href = avatarSrc(logoPath, basePath + '/aibase/favicon.ico');
  document.getElementsByTagName('head')[0].appendChild(link);
};

export const replaceTitle = title => {
  document.title = title || i18n.t('header.title');
};

export const getModelDefaultIcon = () => {
  const { defaultIcon = {} } = store.state.user.commonInfo.data || {};
  return (
    avatarSrc(defaultIcon?.modelIcon) ||
    require('@/assets/imgs/model_default_icon.png')
  );
};

export const copy = text => {
  let textareaEl = document.createElement('textarea');
  textareaEl.setAttribute('readonly', 'readonly'); // 防止手机上弹出软键盘
  textareaEl.value = text;
  document.body.appendChild(textareaEl);
  textareaEl.select();
  const res = document.execCommand('copy');
  textareaEl.remove();
  return res;
};

export const copyCb = () => {
  Message.success(i18n.t('common.copy.success'));
};

// 直链下载
export function directDownload(url, filename = '') {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();
  link.remove();
}

export const resDownloadFile = (response = {}, fileName = '') => {
  const blob = new Blob([response], { type: response.type });
  const url = URL.createObjectURL(blob);
  directDownload(url, fileName);
  URL.revokeObjectURL(url);
};

// fetch请求下载（强制重命名）
export async function fetchDownload(url, filename = '') {
  try {
    const response = await fetch(url);
    if (!response.ok) throw new Error('文件下载失败');
    const blob = await response.blob();
    resDownloadFile(blob, filename);
  } catch (error) {
    console.error('下载出错:', error);
    directDownload(url, filename);
  }
}

// 带鉴权的文件下载（/v1/files/* 路由要求 JWT）
export async function authDownload(url, fileName = '') {
  const token = store.getters['user/token'] || '';
  const user = store.getters['user/userInfo'] || {};
  const response = await fetch(url, {
    headers: {
      Authorization: `Bearer ${token}`,
      'x-user-id': user.uid || '',
      'x-org-id': user.orgId || '',
      'X-Client-ID': getXClientId(),
    },
  });
  if (!response.ok) {
    throw new Error(`下载失败: ${response.status} ${response.statusText}`);
  }
  const blob = await response.blob();
  resDownloadFile(blob, fileName);
}

export function convertLatexSyntax(inputText) {
  // 1. 匹配块级公式，将 `\[` 和 `\]` 替换为 `$$`，支持 `\\[` `\\]` 或单个 `\[` `\]`
  inputText = inputText.replaceAll(
    /\\\[\s*([\s\S]+?)\s*\\\]/g,
    (_, formula) => `$$${formula}$$`,
  );
  // 2. 匹配行内公式，将 `\(` 和 `\)` 替换为 `$`，支持 `\\(` `\\)` 或单个 `\(` `\)`
  inputText = inputText.replaceAll(
    /\\\(\s*([\s\S]+?)\s*\\\)/g,
    (_, formula) => `$${formula}$`,
  );
  return inputText;
}

export function formatTimestamp(timestamp, format = 'YYYY-MM-DD HH:mm:ss') {
  const date = new Date(timestamp);

  const map = {
    YYYY: date.getFullYear(),
    MM: String(date.getMonth() + 1).padStart(2, '0'),
    DD: String(date.getDate()).padStart(2, '0'),
    HH: String(date.getHours()).padStart(2, '0'),
    mm: String(date.getMinutes()).padStart(2, '0'),
    ss: String(date.getSeconds()).padStart(2, '0'),
  };

  return format.replaceAll(/YYYY|MM|DD|HH|mm|ss/g, matched => map[matched]);
}

export function isSub(data) {
  return /【(\d{0,2})\^】/.test(data);
}

export function parseSub(data, index) {
  // 标点吸附：汉字与引用之间、引用与中文/英文标点之间的空白全部压掉，
  // 避免出现 "水平 【1^】 。它..." 这种被空格割裂的阅读节奏
  data = data
    .replaceAll(/([\u4e00-\u9fa5A-Za-z0-9])\s+(?=【\d{0,2}\^】)/g, '$1')
    .replaceAll(/【(\d{0,2})\^】\s+(?=[，。！？；：、）】》」"'])/g, '【$1^】');
  return data.replaceAll(/【(\d{0,2})\^】/g, item => {
    const num = item.match(/【(\d{0,2})\^】/)[1];
    return `<sup class='citation' data-parents-index='${index}'>${num}</sup>`;
  });
}

// 子会话专用的 parseSub
export function parseSubConversation(text, index, searchList, id) {
  return text.replaceAll(/【(\d{0,2})\^】/g, item => {
    let result = item.match(/【(\d{0,2})\^】/)[1];
    return `<sup class='citation' data-parents-index='${index}' data-pid='${id}'>${result}</sup>`;
  });
}

// 是否是有效的URL
export function isValidURL(string) {
  const res = string.match(
    /(https?|ftp|file|ssh):\/\/[-A-Z0-9+&@#/%?=~_|!:,.;]*[-A-Z0-9+&@#/%=~_|]/i,
  );
  return res !== null;
}

export function isExternal(path) {
  return /^(https?:|mailto:|tel:)/.test(path);
}

export const formatTools = tools => {
  if (!tools?.length) return [];
  return tools.map((n, i) => {
    let params = [];
    let properties = n.inputSchema.properties;
    for (let key in properties) {
      params.push({
        name: key,
        requiredBadge: n.inputSchema.required?.includes(key)
          ? i18n.t('common.required')
          : '',
        type: properties[key].type,
        description: properties[key].description,
      });
    }
    return {
      ...n,
      params,
    };
  });
};

/**
 * 格式化得分，保留5位小数
 * @param {number} score - 得分值
 * @returns {string} 格式化后的得分字符串
 */
export function formatScore(score) {
  if (typeof score !== 'number') {
    return '0.00000';
  }
  return score.toFixed(5);
}

export function avatarSrc(path, defaultImg = '') {
  if (!path) return defaultImg;
  if (path.startsWith('http')) return path;
  return basePath + '/user/api/' + path;
}

// 换算单位万/亿/万亿，保留2位小数
export const formatAmount = (
  num,
  returnType = 'string',
  preserveRange = false,
) => {
  if (!num) return 0;

  const units = i18n.t('statisticsEcharts.units');
  const isHasDecimal = num.toString().includes('.');
  let formatNum = num;
  let simplifiedNum = num.toString();

  // 99999以内原样显示
  if (preserveRange && num < 100000) {
    if (returnType === 'object') {
      return {
        value: simplifiedNum,
        type: '',
      };
    } else {
      return simplifiedNum;
    }
  }

  if (isHasDecimal) {
    formatNum = Number(num.toString().slice(0, num.toString().indexOf('.')));
  }
  // 获取数字的数量级
  let unitIndex = Math.floor((String(formatNum).length - 1) / 4);

  if (unitIndex > 0) {
    const unit = units[unitIndex];

    const divisor = Math.pow(10, unitIndex * 4);
    //缩小相应倍数，并保留2位小数
    const formattedValue = (num / divisor)
      .toFixed(2)
      .replaceAll(/(\d)(?=(\d{3})+(?!\d))/g, '$1,');

    if (returnType === 'object') {
      return {
        value: formattedValue,
        type: unit,
      };
    } else {
      simplifiedNum = formattedValue + unit;
    }
  } else if (returnType === 'object') {
    // 数量级为0时的对象格式返回
    return {
      value: simplifiedNum,
      type: '',
    };
  }

  return simplifiedNum;
};

export function deepMerge(obj1, obj2) {
  for (let key in obj2) {
    if (obj2[key] && typeof obj2[key] === 'object') {
      if (!obj1[key] || typeof obj1[key] !== 'object') {
        obj1[key] = {};
      }
      deepMerge(obj1[key], obj2[key]);
    } else {
      obj1[key] = obj2[key];
    }
  }
  return obj1;
}

/**
 * 防抖函数（Debounce）
 * 限制函数在一定时间内的执行频率，合并短时间内的多次调用为一次
 * @param {Function} func - 需要防抖的函数
 * @param {number} wait - 等待时间（毫秒）
 * @param {boolean} immediate - 是否立即执行
 * @returns {Function} 防抖处理后的函数
 */
export function debounce(func, wait, immediate) {
  let timeout, args, context, timestamp, result;

  const later = function () {
    // 计算上次调用时间与当前时间的差值
    const last = Date.now() - timestamp;

    // 如果上次调用时间与当前时间的差值小于wait，则设置新的定时器
    if (last < wait && last >= 0) {
      timeout = setTimeout(later, wait - last);
    } else {
      // 否则执行函数
      timeout = null;
      if (!immediate) {
        result = func.apply(context, args);
        context = args = null;
      }
    }
  };

  return function () {
    context = this;
    args = arguments;
    timestamp = Date.now();

    // 如果immediate为true且当前没有定时器，则立即执行函数
    const callNow = immediate && !timeout;

    // 设置定时器
    if (!timeout) {
      timeout = setTimeout(later, wait);
    }

    // 如果需要立即执行，则立即调用函数
    if (callNow) {
      result = func.apply(context, args);
      context = args = null;
    }

    return result;
  };
}

// 获取文件icon
export function getFileIcon(type) {
  switch (type) {
    case 'txt':
      return require('@/assets/imgs/txt-icon.png');
    case 'csv':
      return require('@/assets/imgs/csv-icon.png');
    case 'xlsx':
      return require('@/assets/imgs/xls-icon.png');
    case 'docx':
      return require('@/assets/imgs/word-icon.png');
    case 'pptx':
      return require('@/assets/imgs/ppt-icon.png');
    case 'pdf':
      return require('@/assets/imgs/pdf-icon.png');
    default:
      return require('@/assets/imgs/fileicon.png');
  }
}

// 根据文件名提取文件类型（配合FileIcon组件使用）
export function getFileIconType(fileOrName) {
  const filename =
    typeof fileOrName === 'string'
      ? fileOrName
      : fileOrName?.name ||
        fileOrName?.fileName ||
        fileOrName?.oldFileName ||
        fileOrName?.oldName ||
        '';
  const pureName = filename.split('?')[0].split('#')[0];
  const ext = (pureName.split('.').pop() || '').toLowerCase();
  const normalizedExtMap = {
    jpeg: 'jpg',
  };
  return normalizedExtMap[ext] || ext || 'unknown';
}

// 文件大小格式化
export function formatFileSize(bytes, decimals = 2) {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return (
    Number.parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) +
    ' ' +
    sizes[i]
  );
}

export function Md2Img(markdownText, escapeHtml = true) {
  // 匹配 Markdown 图片语法的正则表达式
  // ![](image.jpg) 或 ![alt](image.jpg) 或 ![alt](image.jpg "title")
  const imageRegex = /!\[(.*?)\]\(([^)\s]+)(?:\s+"([^"]*)")?\)/g;
  // 匹配 Markdown 换行符的正则表达式
  const newlineRegex = /(\r\n|\r|\n)/g;

  // 转义HTML特殊字符
  if (escapeHtml)
    markdownText = markdownText
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;');

  let lastIndex = 0;
  let result = '';

  let match;
  while ((match = imageRegex.exec(markdownText)) !== null) {
    // 添加匹配前的文本内容
    result += markdownText.substring(lastIndex, match.index);

    const alt = match[1] || '';
    const src = match[2];
    const title = match[3] ? ` title="${sanitizeEscapeHtml(match[3])}"` : '';

    // 校验 URL scheme，阻止 javascript: 等危险协议
    if (isSafeImageUrl(src)) {
      const safeSrc = src.replaceAll('"', '&quot;');
      const safeAlt = alt.replaceAll('"', '&quot;');
      result += `<img src="${safeSrc}" alt="${safeAlt}"${title}>`;
    } else {
      // 不安全的 URL：渲染为纯文本而非图片标签
      result += `![${alt}](${src})`;
    }

    // 更新lastIndex到匹配结束位置
    lastIndex = match.index + match[0].length;
  }

  // 添加剩余的文本内容
  result += markdownText.substring(lastIndex);

  // 将换行符转换为<br>标签
  result = result.replaceAll(newlineRegex, '<br>');

  return result;
}

export function Img2Md(htmlString, escapeHtml = true) {
  if (['<div><br></div>', '<br>'].includes(htmlString)) return '';
  // 匹配 img 标签的正则表达式
  const imgRegex = /<img\s+[^>]*src\s*=\s*["']([^"']+)["'][^>]*>/gi;

  // 替换 img 标签为 Markdown 格式
  let result = htmlString.replaceAll(imgRegex, (match, src) => {
    // 提取 alt 属性（如果有）
    const altMatch = match.match(/alt\s*=\s*["']([^"']*)["']/i);
    const alt = altMatch ? altMatch[1] : '';
    return `![${alt}](${src})`;
  });

  result = result
    // 处理空行
    .replaceAll(/<div><br><\/div>/gi, '\n')
    // 处理块级元素的换行 - 仅在块级元素前添加换行符，后截替换为空
    .replaceAll(/<(div|p|h[1-6]|li|blockquote)\b[^>]*>(.*?)<\/\1>/gi, '\n$2')
    // 处理自闭合的br标签
    .replaceAll(/<br\s*\/?>/gi, '\n')
    // 删除所有其他HTML标签，只保留纯文本内容和换行符
    .replaceAll(/<[^>]*>/g, '');

  // 恢复HTML特殊字符
  if (escapeHtml)
    result = result
      .replaceAll('&lt;', '<')
      .replaceAll('&gt;', '>')
      .replaceAll('&quot;', '"')
      .replaceAll('&#39;', "'")
      .replaceAll('&amp;', '&');

  return result;
}

//   格式化文件大小
export function filterSize(size) {
  if (!size) return '';
  const num = 1024; //byte
  if (size < num) return size + 'B';
  if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + 'KB'; //kb
  if (size < Math.pow(num, 3))
    return (size / Math.pow(num, 2)).toFixed(2) + 'MB'; //M
  if (size < Math.pow(num, 4))
    return (size / Math.pow(num, 3)).toFixed(2) + 'G'; //G
  return (size / Math.pow(num, 4)).toFixed(2) + 'T'; //T
}

/**
 * 节流函数
 * @param {Function} fn - 要节流的函数
 * @param {number} interval - 执行间隔（毫秒）
 * @returns {Function} 节流后的函数
 */
export function throttle(fn, interval = 100) {
  let lastTime = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastTime >= interval) {
      fn.apply(this, args);
      lastTime = now;
    }
  };
}

/**
 * 检测文件类型
 * @param {string} fileName - 文件名
 * @returns {string} 文件类型
 */
export function getFileType(fileName) {
  if (!fileName) return 'text';
  const ext = fileName.split('.').pop().toLowerCase();

  const typeMap = {
    image: ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico'],
    video: ['mp4', 'webm', 'ogg', 'mov', 'm4v', 'avi', 'mkv'],
    audio: ['mp3', 'wav', 'ogg', 'm4a', 'flac', 'aac', 'wma'],
    pdf: ['pdf'],
    ppt: ['ppt', 'pptx'],
    excel: ['csv', 'xls', 'xlsx'],
    word: ['doc', 'docx'],
    ofd: ['ofd'],
    html: ['html', 'htm'],
    markdown: ['md'],
  };

  for (const [type, exts] of Object.entries(typeMap)) {
    if (exts.includes(ext)) {
      return type;
    }
  }

  return 'text';
}

/**
 * 检测是否为图片文件
 * @param {Object} file - 文件对象
 * @returns {boolean} 是否为图片
 */
export function isImageFile(file) {
  const imageTypes = [
    'image/jpeg',
    'image/jpg',
    'image/png',
    'image/gif',
    'image/webp',
    'image/bmp',
  ];
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];

  if (file.type && imageTypes.includes(file.type.toLowerCase())) {
    return true;
  }

  const fileName = file.name || file.fileName || '';
  if (fileName) {
    const ext = fileName.split('.').pop().toLowerCase();
    return imageExts.includes(ext);
  }

  return false;
}

/**
 * 根据文件扩展名获取对应的 MIME 类型
 * @param {string} fileExt - 文件扩展名（不含点，如 'png', 'pdf'）
 * @returns {string} MIME 类型字符串，未知类型返回空字符串
 */
export function getMimeType(fileExt) {
  if (!fileExt) return '';

  const ext = fileExt.toLowerCase();

  // MIME 类型映射表
  const mimeTypeMap = {
    // 图片类型
    png: 'image/png',
    jpg: 'image/jpeg',
    jpeg: 'image/jpeg',
    gif: 'image/gif',
    svg: 'image/svg+xml',
    webp: 'image/webp',
    bmp: 'image/bmp',
    ico: 'image/x-icon',
    // 视频类型
    mp4: 'video/mp4',
    webm: 'video/webm',
    ogg: 'video/ogg',
    mov: 'video/quicktime',
    m4v: 'video/x-m4v',
    avi: 'video/x-msvideo',
    mkv: 'video/x-matroska',
    // 音频类型
    mp3: 'audio/mpeg',
    wav: 'audio/wav',
    oga: 'audio/ogg',
    m4a: 'audio/mp4',
    flac: 'audio/flac',
    aac: 'audio/aac',
    wma: 'audio/x-ms-wma',
    // 文档类型
    pdf: 'application/pdf',
    ofd: 'application/ofd',
    html: 'text/html',
    htm: 'text/html',
  };

  return mimeTypeMap[ext] || `text/${ext}`;
}

/**
 * 格式化持续时间
 * @param {number} ms - 毫秒数
 * @returns {string} 格式化后的时间字符串（如 "2m 30s" 或 "500ms"）
 */
export function formatDuration(ms) {
  if (ms === 0) {
    return '<1s';
  }
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  }
  return `${secs}s`;
}

/**
 * 格式化数字，将大数字转换为带 'k' 或 'w' 单位的字符串（采用向下截取法，不进行四舍五入）。
 * * @description
 * - 小于 1000: 返回原数字字符串
 * - 1000 ~ 9999: 转换为 'k' 单位 (如 9999 -> '9.99k')
 * - 10000 及以上: 转换为 'w' 单位 (如 10500 -> '1.05w')
 *
 * @param {number} num - 需要格式化的原始数字。如果传入非数字或 NaN，将统一返回 '0'。
 * @param {number} [fixed=2] - 保留的小数位数。默认为 2。
 * @param {boolean} [strictFixed=true] - 是否严格保留指定的小数位数。
 * - true: 严格补零（如 '1.00k', '1.10w'）
 * - false: 省略末尾无用的零（如 '1k', '1.1w'）
 * @returns {string} 格式化后的数字字符串。
 */
export function formatCount(num, fixed = 2, strictFixed = true) {
  if (typeof num !== 'number' || Number.isNaN(num)) {
    return '0';
  }

  // 小于 1000 的情况，直接原样返回
  if (num < 1000) {
    return String(num);
  }

  // 不四舍五入的向下截取
  const floorToFixed = (value, precision) => {
    const multiplier = Math.pow(10, precision);

    const truncated = Math.floor(value * multiplier) / multiplier;

    const formatted = truncated.toFixed(precision);
    return strictFixed ? formatted : Number(formatted).toString();
  };

  // 3. 大于等于 10000 转换为 'w'
  if (num >= 10000) {
    return `${floorToFixed(num / 10000, fixed)}w`;
  }

  // 4. 1000 ~ 10000 之间转换为 'k'
  return `${floorToFixed(num / 1000, fixed)}k`;
}
