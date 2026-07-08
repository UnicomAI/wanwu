import { i18n } from '@/lang';

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
