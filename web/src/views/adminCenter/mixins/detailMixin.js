import { avatarSrc, getModelDefaultIcon } from '@/utils/util';

const PUBLISH_STATUS_KEY = 'adminCenter.options.publishStatus';

export default {
  computed: {
    backTitle() {
      return (
        this.$t(this.moduleTitleKey) + this.$t('adminCenter.common.backend')
      );
    },
    statusText() {
      const published = this.$t(`${PUBLISH_STATUS_KEY}.published`);
      const map = {
        draft: this.$t(`${PUBLISH_STATUS_KEY}.draft`),
        publish: published,
        published,
      };
      return map[this.base.publishStatus] ?? this.base.publishStatus ?? '-';
    },
    visibleUsers() {
      return (this.base.authorizedPersonnelList || []).map(u => ({
        id: u.userId,
        name: u.name,
        avatar: avatarSrc(u.avatar?.path),
        orgPath: u.orgName,
      }));
    },
    basicItems() {
      return this.baseBasicItems();
    },
  },
  methods: {
    avatarSrc,
    baseBasicItems() {
      const b = this.base;
      return [
        {
          label: this.$t('adminCenter.common.creator'),
          value: b.ownerUserName,
        },
        { label: this.$t('adminCenter.common.org'), value: b.ownerOrgName },
        {
          label: this.$t('adminCenter.common.updateTime'),
          value: b.updatedAt,
        },
      ];
    },
    convertModelIcon(iconPath) {
      return iconPath ? avatarSrc(iconPath) : getModelDefaultIcon();
    },
    numText(v) {
      return v === undefined || v === null || v === '' ? '-' : String(v);
    },
    boolText(v) {
      return v
        ? this.$t('adminCenter.common.enabled')
        : this.$t('adminCenter.common.disabled');
    },
  },
};
