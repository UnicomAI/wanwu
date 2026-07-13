import { i18n } from '@/lang';

export const CHAT = 'chatflow';
export const WORKFLOW = 'workflow';
export const RAG = 'rag';
export const AGENT = 'agent';
export const SKILL = 'skill';
export const WGA = 'wga'; // 通用智能体
export const MODEL = 'model';
export const KNOWLEDGE = 'knowledge';
export const MCP = 'mcp';
export const TOOL = 'tool';
export const PROMPT = 'prompt';
export const SAFETY = 'safety';
export const AppType = {
  [WORKFLOW]: i18n.t('appSpace.workflow'),
  [CHAT]: i18n.t('appSpace.chat'),
  [RAG]: i18n.t('appSpace.rag'),
  [AGENT]: i18n.t('appSpace.agent'),
};
export const WorkflowTypeList = [
  { value: WORKFLOW, name: i18n.t('appSpace.workflow') },
  { value: CHAT, name: i18n.t('appSpace.chat') },
];
export const TotalTypeObj = {
  [AGENT]: i18n.t('appSpace.agent'),
  [WORKFLOW]: i18n.t('appSpace.workflow'),
  [RAG]: i18n.t('appSpace.rag'),
  [WGA]: i18n.t('appSpace.generalAgent'),
  [MODEL]: i18n.t('appSpace.model'),
  [KNOWLEDGE]: i18n.t('appSpace.knowledge'),
  [MCP]: i18n.t('appSpace.mcp'),
  [TOOL]: i18n.t('appSpace.tool'),
  [PROMPT]: i18n.t('appSpace.prompt'),
  [SKILL]: i18n.t('appSpace.skill'),
  [SAFETY]: i18n.t('appSpace.safety'),
};
export const ShowSelectAppList = [KNOWLEDGE, RAG, WORKFLOW, AGENT];
export const TagColorObj = {
  [AGENT]: 'tag-purple',
  [WORKFLOW]: 'tag-green',
  [RAG]: 'tag-blue',
  [CHAT]: 'tag-orange',
  [WGA]: 'tag-purple',
  [MODEL]: 'tag-green',
  [KNOWLEDGE]: 'tag-blue',
  [MCP]: 'tag-orange',
  [TOOL]: 'tag-purple',
  [PROMPT]: 'tag-green',
  [SKILL]: 'tag-blue',
  [SAFETY]: 'tag-orange',
};
export const SafetyType = {
  Political: i18n.t('common.safetyType.political'),
  Revile: i18n.t('common.safetyType.revile'),
  Pornography: i18n.t('common.safetyType.pornography'),
  ViolentTerror: i18n.t('common.safetyType.violentTerror'),
  Illegal: i18n.t('common.safetyType.illegal'),
  InformationSecurity: i18n.t('common.safetyType.informationSecurity'),
  Other: i18n.t('common.safetyType.other'),
};
