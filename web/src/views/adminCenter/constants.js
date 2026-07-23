import { i18n } from '@/lang';

export const ALL_VALUE = 'all';

export const DRAFT = 'draft';
export const PUBLISHED = 'publish';

export const PRIVATE = 'private';
export const ORGANIZATION = 'organization';
export const PUBLIC = 'public';

export const PUBLISH_STATUS_LIST = [
  {
    label: i18n.t('adminCenter.options.publishStatus.draft'),
    value: DRAFT,
  },
  {
    label: i18n.t('adminCenter.options.publishStatus.published'),
    value: PUBLISHED,
  },
];

export const SCOPE_TYPE_LIST = [
  {
    label: i18n.t('adminCenter.options.publishScope.private'),
    value: PRIVATE,
  },
  {
    label: i18n.t('adminCenter.options.publishScope.organization'),
    value: ORGANIZATION,
  },
  {
    label: i18n.t('adminCenter.options.publishScope.public'),
    value: PUBLIC,
  },
];

export const normalizeSelectValues = values => {
  if (!Array.isArray(values) || values.includes(ALL_VALUE)) {
    return [];
  }
  return values.filter(Boolean);
};

export const normalizeAdminListParams = params => {
  const normalized = { ...params };

  [
    'orgIdList',
    'userIdList',
    'publishScope',
    'publishStatus',
    'category',
    'modelType',
    'type',
    'appType',
  ].forEach(key => {
    const values = normalizeSelectValues(normalized[key]);
    if (values.length) {
      normalized[key] = values;
    } else {
      delete normalized[key];
    }
  });

  Object.keys(normalized).forEach(key => {
    if (
      normalized[key] === '' ||
      normalized[key] === undefined ||
      normalized[key] === null
    ) {
      delete normalized[key];
    }
  });

  return normalized;
};
