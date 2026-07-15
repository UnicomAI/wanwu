<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="el-icon-arrow-left" @click="$router.go(-1)" />
      <img class="page-title-img" src="@/assets/imgs/org.png" alt="" />
      <span class="page-title-name">{{ $t('menu.setting') }}</span>
    </div>
    <!-- tabs: 改版，组织、用户、角色，提到一级 tab 标签，不分二级 tab -->
    <div class="tabs tabs-spacing">
      <div
        v-for="item in list"
        v-if="checkPerm(item.perm)"
        :key="item.tab"
        :class="['tab', { active: tabActive === item.tab }]"
        @click="tabClick(item.tab)"
      >
        {{ item.name }}
      </div>
    </div>

    <div v-if="tabActive === 0" class="org-wrapper">
      <Org />
    </div>
    <div v-if="tabActive === 1" class="org-wrapper">
      <User />
    </div>
    <div v-if="tabActive === 2" class="org-wrapper">
      <Role />
    </div>
    <div v-if="tabActive === 3" class="info-setting-wrapper">
      <InfoSetting />
    </div>
    <div v-if="tabActive === 4">
      <Oauth />
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

export default {
  name: 'Permission',
  components: { User, Role, Org, InfoSetting, Oauth },
  data() {
    return {
      tabActive: 0,
      list: [
        {
          name: this.$t('org.title'),
          tab: 0,
        },
        {
          name: this.$t('user.title'),
          tab: 1,
        },
        {
          name: this.$t('role.title'),
          tab: 2,
        },
        {
          name: this.$t('infoSetting.title'),
          tab: 3,
          perm: PERMS.SETTING,
        },
        {
          name: this.$t('oauth.title'),
          tab: 4,
          perm: PERMS.OAUTH,
        },
      ],
    };
  },
  methods: {
    checkPerm,
    tabClick(status) {
      this.tabActive = status;
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';
.page-wrapper {
  .tabs-spacing {
    padding-top: 14px;
    padding-bottom: 10px;
  }
}
.org-wrapper {
  margin: 10px 0 0 20px;
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
  margin: 10px 10px 0 20px;
  max-height: calc(100vh - 170px);
  overflow-y: auto;
  padding-right: 10px;
}
</style>
