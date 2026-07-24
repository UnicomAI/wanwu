<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="el-icon-arrow-left" @click="$router.go(-1)" />
      <img class="page-title-img" src="@/assets/imgs/org.png" alt="" />
      <span class="page-title-name">{{ $t('adminCenter.title') }}</span>
    </div>

    <div v-if="canViewCurrent" :class="currentView.wrapperClass">
      <component :is="currentView.component" />
    </div>
    <div v-else class="no-page-permission">
      {{ $t('common.message.noPagePermission') }}
    </div>
  </div>
</template>

<script>
import User from './user/index.vue';
import Role from './role/index.vue';
import Org from './org/index.vue';
import InfoSetting from '@/views/infoSetting/index.vue';
import Oauth from './oauth/index.vue';
import { checkPerm, PERMS } from '@/router/permission';

const VIEW_MAP = {
  org: {
    component: Org,
    wrapperClass: 'org-wrapper',
    title: 'org.title',
  },
  user: {
    component: User,
    wrapperClass: 'org-wrapper',
    title: 'user.title',
  },
  role: {
    component: Role,
    wrapperClass: 'org-wrapper',
    title: 'role.title',
  },
  platformConfig: {
    component: InfoSetting,
    wrapperClass: 'info-setting-wrapper',
    perm: PERMS.SETTING,
    title: 'infoSetting.title',
  },
  oAuth: {
    component: Oauth,
    perm: PERMS.OAUTH,
    wrapperClass: 'oauth-setting-wrapper',
    title: 'oauth.title',
  },
};
export default {
  name: 'Permission',
  components: { User, Role, Org, InfoSetting, Oauth },
  computed: {
    currentView() {
      return VIEW_MAP[this.$route.query.key] || VIEW_MAP.org;
    },
    canViewCurrent() {
      return this.checkPerm(this.currentView.perm);
    },
  },
  methods: {
    checkPerm,
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';
.page-wrapper {
  height: calc(100vh - 32px);
  display: flex;
  flex-direction: column;
  .tabs-spacing {
    padding-top: 14px;
    padding-bottom: 10px;
  }
}
.org-wrapper {
  padding: 10px 0 14px 20px;
  height: 100%;
  overflow: hidden;
}
.page-title {
  .el-icon-arrow-left {
    margin-right: 10px;
    font-size: 15px;
    cursor: pointer;
    color: $color_title;
  }
}
.info-setting-wrapper {
  margin: 10px 10px 10px 20px;
  max-height: calc(100% - 20px);
  overflow-y: auto;
}
.no-page-permission {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 150px);
  color: #ccc;
}
.oauth-setting-wrapper {
  flex: 1;
  overflow: hidden;
}
</style>
