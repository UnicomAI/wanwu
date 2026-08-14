<template>
  <div class="page-wrapper">
    <div class="dashboard-filter-bar dashboard-sticky-header">
      <div class="global-filter-section">
        <GlobalFilter
          v-if="isShowGlobal"
          ref="globalFilter"
          @change="handleGlobalFilterChange"
        />
      </div>
      <div class="search-section">
        <Search ref="search" @handleSetTime="handleSearchTime" />
      </div>
    </div>

    <div class="dashboard-dimension-bar">
      <div class="dimension-cards">
        <div
          v-for="item in dimensionList"
          :key="item.type"
          :class="['dimension-card', { active: activeDimension === item.type }]"
          @click="activeDimension = item.type"
        >
          <div class="dimension-icon">
            <img :src="item.icon" alt="" />
          </div>
          <div class="dimension-info">
            <span class="dimension-title">{{ item.name }}</span>
            <span class="dimension-desc">{{ item.desc }}</span>
          </div>
        </div>
      </div>

      <div class="tabs" v-if="activeDimension !== STATISTIC.API">
        <div
          v-for="item in scopeList"
          :key="item.type"
          :class="['tab', { active: activeScope === item.type }]"
          @click="activeScope = item.type"
        >
          {{ item.name }}
        </div>
      </div>
    </div>

    <div class="dashboard-content">
      <Model
        v-if="activeDimension === STATISTIC.MODEL"
        :global-filter-params="globalFilterParams"
        :time-params="searchTime"
        :scope="activeScope"
      />
      <App
        v-if="activeDimension === STATISTIC.APP"
        :global-filter-params="globalFilterParams"
        :time-params="searchTime"
        :scope="activeScope"
      />
      <API
        v-if="activeDimension === STATISTIC.API"
        :global-filter-params="globalFilterParams"
        :time-params="searchTime"
      />
    </div>
  </div>
</template>

<script>
import Model from './components/model/model.vue';
import App from './components/app/app.vue';
import API from './components/api/api.vue';
import GlobalFilter from './components/globalFilter.vue';
import Search from '@/components/searchRangeDate.vue';
import { STATISTIC, SCOPE, ALL } from './constants';

export default {
  name: 'StatisticsDashboard',
  components: { Model, App, API, GlobalFilter, Search },
  data() {
    const { isSystem, isAdmin } = this.$store.state.user.permission || {};
    return {
      STATISTIC,
      SCOPE,
      ALL,
      isShowGlobal: isSystem || isAdmin,
      activeDimension: STATISTIC.MODEL,
      activeScope: SCOPE.PUBLISHED,
      globalFilterParams:
        isSystem || isAdmin ? { orgIds: [ALL], userIds: [ALL] } : {},
      searchTime: { time: [] },
      dimensionList: [
        {
          type: STATISTIC.MODEL,
          name: this.$t('statisticsDashboard.modelDimension'),
          desc: this.$t('statisticsDashboard.modelDimensionDesc'),
          icon: require('@/assets/imgs/dashboard_model.png'),
        },
        {
          type: STATISTIC.APP,
          name: this.$t('statisticsDashboard.appDimension'),
          desc: this.$t('statisticsDashboard.appDimensionDesc'),
          icon: require('@/assets/imgs/dashboard_app.png'),
        },
        {
          type: STATISTIC.API,
          name: this.$t('statisticsDashboard.apiDimension'),
          desc: this.$t('statisticsDashboard.apiDimensionDesc'),
          icon: require('@/assets/imgs/dashboard_api.png'),
        },
      ],
      scopeList: [
        {
          type: SCOPE.PUBLISHED,
          name: this.$t('statisticsDashboard.published'),
        },
        {
          type: SCOPE.USED,
          name: this.$t('statisticsDashboard.used'),
        },
      ],
    };
  },
  methods: {
    handleGlobalFilterChange(vals) {
      this.globalFilterParams = { ...vals };
    },
    handleSearchTime(val) {
      this.searchTime = val;
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';

.page-wrapper {
  position: relative;
  height: calc(100vh - 32px);
  overflow-y: auto;
  padding: 0 0 16px 0;
}

.dashboard-sticky-header {
  position: sticky;
  top: 0;
  z-index: 200;
  background: #fff;
  box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.06);
}

.dashboard-filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  min-height: 60px;
}

.global-filter-section {
  flex: 1;

  ::v-deep .global-filter-wrapper {
    padding: 0 24px;
  }
}

.search-section {
  flex: 0 0 auto;

  ::v-deep .statistics_search_time {
    width: auto !important;
    padding: 0 24px;
  }
}

.dashboard-dimension-bar {
  padding: 16px 24px 0;
}

.dimension-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 10px;
}

.dimension-card {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 16px;
  background: #f7f8fa;
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: #f0f2f5;
  }

  &.active {
    background: #fff;
    border-color: $color;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);
  }
}

.dimension-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 8px;
  margin-right: 12px;
  flex-shrink: 0;

  img {
    width: 100%;
    height: 100%;
    border-radius: 8px;
  }
}

.dimension-info {
  display: flex;
  flex-direction: column;
}

.dimension-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.dimension-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.dashboard-content {
  padding: 16px 0;

  ::v-deep .statistics_common {
    height: auto !important;
    overflow: visible !important;
  }

  ::v-deep .statistics_content_box {
    height: auto !important;
    overflow: visible !important;
  }
}
</style>
