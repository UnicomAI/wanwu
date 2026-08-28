/**
 * 对话日志来源枚举。
 * 值与日志列表接口的 source 字段保持一致，供筛选与展示转换复用。
 */
export const CONVERSATION_LOG_SOURCE = Object.freeze({
  WEB: 'web',
  OPEN_API: 'openapi',
  WEB_URL: 'webURL',
  DRAFT: 'draft',
});

/**
 * 对话日志来源筛选项。
 * labelKey 对应 agent.log.sourceOptions 下的国际化文案。
 */
export const CONVERSATION_LOG_SOURCE_OPTIONS = Object.freeze([
  {
    value: CONVERSATION_LOG_SOURCE.WEB,
    labelKey: 'web',
  },
  {
    value: CONVERSATION_LOG_SOURCE.OPEN_API,
    labelKey: 'openApi',
  },
  {
    value: CONVERSATION_LOG_SOURCE.WEB_URL,
    labelKey: 'webUrl',
  },
  {
    value: CONVERSATION_LOG_SOURCE.DRAFT,
    labelKey: 'draft',
  },
]);
