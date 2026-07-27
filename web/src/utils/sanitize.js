import DOMPurify from 'dompurify';

/**
 * DOMPurify 配置
 *
 * ADD_TAGS: 应用自定义标签 <think> 和 <tool>
 *   - 这两个非标准标签通过 markdown-it (html:true) 传递，
 *     在 streamMessageField.vue 的 replaceHTML() 中被替换为 <section>
 *   - DOMPurify 默认会剥离未知标签，必须显式允许
 *
 * ALLOW_DATA_ATTR: true
 *   - 应用使用 data-* 属性实现引用导航、复制按钮、子会话标识等功能
 *   - data-* 属性无法执行 JavaScript，允许是安全的
 *   - 涉及属性：data-clipboard-text, data-citation-index, data-sub-id 等
 */
const PURIFY_CONFIG = {
  ADD_TAGS: ['think', 'tool'],
  ALLOW_DATA_ATTR: true,
};

/**
 * 对 HTML 进行净化，防止 XSS 攻击同时保留合法内容
 * @param {string} html - 待净化的 HTML 字符串
 * @returns {string} 净化后的安全 HTML
 */
export function sanitizeHtml(html) {
  if (typeof html !== 'string') return '';
  return DOMPurify.sanitize(html, PURIFY_CONFIG);
}

/**
 * 净化 HTML 并剥离所有超链接 <a> 标签，保留链接文字
 * 在 sanitizeHtml 配置基础上增加 FORBID_TAGS:['a']，保证其它净化行为一致
 * @param {string} html - 待净化的 HTML
 * @returns {string} 净化后不含 <a> 的安全 HTML
 */
export function stripLinks(html) {
  if (typeof html !== 'string') return '';
  return DOMPurify.sanitize(html, { ...PURIFY_CONFIG, FORBID_TAGS: ['a'] });
}

/**
 * 转义 HTML 特殊字符，用于纯文本场景（如 parseTxt）
 * 性能优于 DOMPurify，适用于不需要保留 HTML 标签的场景
 * @param {string} text - 纯文本
 * @returns {string} 转义后的安全字符串
 */
export function escapeHtml(text) {
  if (typeof text !== 'string') return '';
  return text
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

/**
 * 安全的 parseTxt 替代方法：先转义 HTML 特殊字符，再转换空白符
 * 用于替代各组件中重复定义的 parseTxt 方法
 * @param {string} txt - 包含换行符和制表符的纯文本
 * @returns {string} 安全的 HTML 字符串
 */
export function parseTxtSafe(txt) {
  return escapeHtml(txt)
    .replaceAll('\n\t', '<br/>&nbsp;')
    .replaceAll('\n', '<br/>')
    .replaceAll('\t', '   &nbsp;');
}

/**
 * 校验图片 URL scheme 是否安全
 * 仅允许 http/https、data:image、根相对路径、当前目录相对路径
 * @param {string} url - 待校验的 URL
 * @returns {boolean} 是否为安全的 URL
 */
export function isSafeImageUrl(url) {
  if (typeof url !== 'string') return false;
  return /^(https?:\/\/|data:image\/|\/|\.\/)/i.test(url);
}
