<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="el-icon-arrow-left" @click="$router.go(-1)" />
      <img
        class="page-title-img"
        src="@/assets/imgs/operationManage.svg"
        alt=""
      />
      <span class="page-title-name">{{ $t('menu.operationManage') }}</span>
    </div>
    <div
      class="tabs tabs-x-top"
      style="margin-bottom: -20px"
      v-if="checkPerm(operationPerm)"
    >
      <div
        :class="['tab', { active: tabActive === 1 }]"
        @click="tabClick(1)"
        v-if="checkPerm(statisticsPerm)"
      >
        {{ $t('statistics.title') }}
      </div>
      <!-- 渠道配置无单独权限，有运营管理权限则有渠道配置权限 -->
      <div :class="['tab', { active: tabActive === 2 }]" @click="tabClick(2)">
        {{ $t('channel.title') }}
      </div>
    </div>

    <div v-if="tabActive === 1" style="margin: 30px 20px 0 20px">
      <Statistics />
    </div>
    <div v-if="tabActive === 2" style="margin: 0 20px">
      <Channel />
    </div>
  </div>
</template>

<script>
import Statistics from '@/views/permission/statistics';
import Oauth from '@/views/permission/oauth';
import Channel from './channel';
import { checkPerm, PERMS } from '@/router/permission';

export default {
  name: 'Operation',
  components: { Statistics, Oauth, Channel },
  data() {
    return {
      radio: '',
      tabActive: 1,
      operationPerm: PERMS.OPERATION,
      statisticsPerm: PERMS.STATISTIC,
    };
  },
  created() {
    if (checkPerm(this.statisticsPerm)) {
      this.tabActive = 1;
    } else {
      this.tabActive = 2;
    }
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
.page-title {
  .el-icon-arrow-left {
    margin-right: 10px;
    font-size: 15px;
    cursor: pointer;
    color: $color_title;
  }
}
</style>
