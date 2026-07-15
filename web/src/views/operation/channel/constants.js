import { AGENT } from '@/utils/commonSet';
import { i18n } from '@/lang';

export const WECHAT = 'wechat';
export const DING_TALK = 'dingtalk';
export const GENERAL_AGENT = 'wga';
export const DIGITAL_EMPLOYEE = 'dip';

export const APP_TYPE_OPTIONS = [
  { value: AGENT, label: i18n.t('channel.agent') },
  { value: GENERAL_AGENT, label: i18n.t('channel.generalAgent') },
  { value: DIGITAL_EMPLOYEE, label: i18n.t('channel.digitalEmployee') },
];
