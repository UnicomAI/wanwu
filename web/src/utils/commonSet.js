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
export const DIGITAL_EMPLOYEE = 'digitalemployee';
export const AppType = {
  [WORKFLOW]: i18n.t('appSpace.workflow'),
  [CHAT]: i18n.t('appSpace.chat'),
  [RAG]: i18n.t('appSpace.rag'),
  [AGENT]: i18n.t('appSpace.agent'),
  [DIGITAL_EMPLOYEE]: i18n.t('appSpace.digitalEmployee'),
};
export const WorkflowTypeList = [
  { value: WORKFLOW, name: i18n.t('appSpace.workflow') },
  { value: CHAT, name: i18n.t('appSpace.chat') },
];
export const TotalTypeObj = {
  [AGENT]: i18n.t('appSpace.agent'),
  [WORKFLOW]: i18n.t('appSpace.workflow'),
  [CHAT]: i18n.t('appSpace.chat'),
  [RAG]: i18n.t('appSpace.rag'),
  [DIGITAL_EMPLOYEE]: i18n.t('appSpace.digitalEmployee'),
  [WGA]: i18n.t('appSpace.generalAgent'),
  [MODEL]: i18n.t('appSpace.model'),
  [KNOWLEDGE]: i18n.t('appSpace.knowledge'),
  [MCP]: i18n.t('appSpace.mcp'),
  [TOOL]: i18n.t('appSpace.tool'),
  [PROMPT]: i18n.t('appSpace.prompt'),
  [SKILL]: i18n.t('appSpace.skill'),
  [SAFETY]: i18n.t('appSpace.safety'),
};
export const ShowSelectAppList = [
  KNOWLEDGE,
  RAG,
  WORKFLOW,
  AGENT,
  DIGITAL_EMPLOYEE,
];
export const TagColorObj = {
  [AGENT]: 'tag-purple',
  [WORKFLOW]: 'tag-green',
  [RAG]: 'tag-blue',
  [CHAT]: 'tag-orange',
  [DIGITAL_EMPLOYEE]: 'tag-purple',
  [WGA]: 'tag-purple',
  [MODEL]: 'tag-cyan',
  [KNOWLEDGE]: 'tag-yellow',
  [MCP]: 'tag-orange',
  [TOOL]: 'tag-purple',
  [PROMPT]: 'tag-green',
  [SKILL]: 'tag-blue',
  [SAFETY]: 'tag-ltGreen',
};
export const TagColorList = [
  { color: '#3562E7', backgroundColor: '#E6F0FF' },
  { color: '#00A56E', backgroundColor: 'rgba(92, 192, 103, 0.15)' },
  { color: '#E87B00', backgroundColor: '#FFF3E5' },
  { color: '#0DA5A5', backgroundColor: '#E7F7F7' },
  { color: '#6349E8', backgroundColor: '#F1EDFF' },
  { color: '#67C23A', backgroundColor: '#F0F9EB' },
  { color: '#E6A23C', backgroundColor: '#FDF6EC' },
];
export const OrgTagColorList = [
  { color: '#3562E7', backgroundColor: '#E6F0FF' },
  { color: '#6349E8', backgroundColor: '#F1EDFF' },
  { color: '#00A56E', backgroundColor: 'rgba(92, 192, 103, 0.15)' },
  { color: '#E87B00', backgroundColor: '#FFF3E5' },
  { color: '#0DA5A5', backgroundColor: '#E7F7F7' },
];
export const SafetyType = {
  Political: i18n.t('common.safetyType.political'),
  Revile: i18n.t('common.safetyType.revile'),
  Pornography: i18n.t('common.safetyType.pornography'),
  ViolentTerror: i18n.t('common.safetyType.violentTerror'),
  Illegal: i18n.t('common.safetyType.illegal'),
  InformationSecurity: i18n.t('common.safetyType.informationSecurity'),
  Other: i18n.t('common.safetyType.other'),
};
