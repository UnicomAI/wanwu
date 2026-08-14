import { i18n } from '@/lang';
import {
  ASR,
  EMBEDDING,
  GUI,
  LLM,
  MULTIMODAL_EMBEDDING,
  MULTIMODAL_RERANK,
  OCR,
  PDF_PARSER,
  RERANK,
} from '@/views/modelAccess/constants';

export const STATISTIC = {
  APP: 'app',
  MODEL: 'model',
  API: 'api',
};

export const SCOPE = {
  PUBLISHED: 'published',
  USED: 'used',
};

export const ALL = 'ALL';
export const DEFAULT_APP_ITEM = {
  keyId: ALL,
  name: i18n.t('statisticsDashboard.all'),
};

export const MODEL_TAG_COLOR = {
  [LLM]: 'tag-blue',
  [RERANK]: 'tag-orange',
  [EMBEDDING]: 'tag-green',
  [OCR]: 'tag-purple',
  [ASR]: 'tag-yellow',
  [GUI]: 'tag-cyan',
  [PDF_PARSER]: 'tag-ltGreen',
  [MULTIMODAL_RERANK]: 'tag-orange',
  [MULTIMODAL_EMBEDDING]: 'tag-green',
};
