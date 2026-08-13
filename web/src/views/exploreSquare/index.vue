<template>
  <div class="page-wrapper">
    <div class="app-header">
      <div class="header-top">
        <div class="taglist_warp">
          <div
            v-for="item in tagList"
            class="tagList"
            @click="handleTagClick(item)"
            :class="{ white: item.value === active }"
          >
            <img
              :src="item.value === active ? item.activeImg : item.unactiveImg"
              class="h-icon"
            />
            <span>{{ item.name }}</span>
          </div>
        </div>
        <SearchInput
          :placeholder="placeholder"
          style="width: 200px"
          @handleSearch="handleSearch"
        />
      </div>
      <div class="explore-tab-pane">
        <el-tabs v-model="activeName" @tab-click="handleClick">
          <el-tab-pane
            v-for="item in appList"
            :key="item.type"
            :label="item.name"
            :name="item.type"
          >
            <AppList
              :appData="listData"
              :isShowTool="false"
              :appFrom="'explore'"
            />
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script>
import SearchInput from '@/components/searchInput.vue';
import AppList from '@/components/appList.vue';
import CreateTotalDialog from '@/components/createTotalDialog.vue';
import { getExplorationList } from '@/api/explore';
import {
  AGENT,
  WORKFLOW,
  RAG,
  CHAT,
  DIGITAL_EMPLOYEE,
} from '@/utils/commonSet';

export default {
  name: 'ExploreSquare',
  components: { SearchInput, CreateTotalDialog, AppList },
  data() {
    return {
      placeholder: this.$t('appSpace.search'),
      asideTitle: this.$t('explore.asideTitle'),
      activeName: 'agent',
      searchValue: '',
      active: 'all',
      tagList: [
        {
          name: this.$t('explore.tag.all'),
          value: 'all',
          activeImg: require('@/assets/imgs/all_active.svg'),
          unactiveImg: require('@/assets/imgs/all_unactive.svg'),
        },
        {
          name: this.$t('explore.tag.favorite'),
          value: 'favorite',
          activeImg: require('@/assets/imgs/mine_active.svg'),
          unactiveImg: require('@/assets/imgs/mine_unactive.svg'),
        },
        {
          name: this.$t('explore.tag.private'),
          value: 'private',
          activeImg: require('@/assets/imgs/start_active.svg'),
          unactiveImg: require('@/assets/imgs/start_unactive.svg'),
        },
        {
          name: this.$t('explore.tag.history'),
          value: 'history',
          activeImg: require('@/assets/imgs/history_active.svg'),
          unactiveImg: require('@/assets/imgs/history_unactive.svg'),
        },
      ],
      historyList: [],
      listData: [],
      appList: [
        { name: this.$t('menu.app.agent'), type: AGENT },
        { name: this.$t('menu.app.rag'), type: RAG },
        { name: this.$t('menu.app.workflow'), type: WORKFLOW },
        { name: this.$t('menu.app.chatflow'), type: CHAT },
        { name: this.$t('menu.app.digitalEmployee'), type: DIGITAL_EMPLOYEE },
      ],
    };
  },
  watch: {
    historyAppList: {
      handler(val) {
        if (val) {
          this.historyList = val;
        }
      },
    },
  },
  created() {
    const { type } = this.$route.query || {};
    this.activeName = [WORKFLOW, RAG, CHAT, DIGITAL_EMPLOYEE].includes(type)
      ? type
      : AGENT;
    this.getExplorationList(this.activeName, this.active);
  },
  mounted() {},
  methods: {
    handleSearch(value) {
      this.searchValue = value;
      this.getExplorationList(this.activeName, this.active);
    },
    historyClick(n) {
      if (!n.path) return;
      this.$router.push({ path: n.path });
    },
    handleClick() {
      this.getExplorationList(this.activeName, this.active);
      if (this.activeName === AGENT) {
        this.$router.replace({ query: {} });
      } else {
        this.$router.replace({ query: { type: this.activeName } });
      }
    },
    handleTagClick(item) {
      this.active = item.value;
      this.getExplorationList(this.activeName, this.active);
    },
    getExplorationList(appType, searchType) {
      const data = { name: this.searchValue, appType, searchType };
      getExplorationList(data).then(res => {
        if (res.code === 0) {
          this.listData = res.data.list || [];
        }
      });
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/tabs.scss';
::v-deep {
  .el-tabs__content {
    overflow: unset;
  }

  .table-search-input {
    height: 30px;
  }
}

.white {
  font-weight: bold;
  color: $color;
  border-bottom: 2.5px solid $color !important;
}

.appList:hover {
  background-color: $color_opacity !important;
}

.appList {
  margin: 10px 20px;
  padding: 10px;
  border-radius: 6px;
  margin-bottom: 6px;
  display: flex;
  gap: 8px;
  align-items: center;
  cursor: pointer;

  .appImg {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    object-fit: cover;
  }

  .appName {
    display: block;
    max-width: 130px;
    overflow: hidden;
    white-space: nowrap;
    pointer-events: none;
    text-overflow: ellipsis;
  }
}

.page-wrapper {
  padding: 10px 22px 20px 30px;
  box-sizing: border-box;

  ::v-deep .app-card {
    padding-right: 8px;
  }

  .header-top {
    display: flex;
    justify-content: space-between;
    padding: 15px 0 6px 0;
    box-sizing: border-box;

    .tagList:nth-child(1) {
      margin-left: 0 !important;
    }

    .taglist_warp {
      display: flex;
      margin-top: -20px;

      .tagList {
        margin: 10px;
        padding: 0 3px;
        height: 36px;
        line-height: 36px;
        cursor: pointer;
        display: flex;
        align-items: center;
        border-bottom: 2.5px solid rgba(255, 255, 255, 0);

        .h-icon {
          margin-right: 5px;
          width: 14px;
        }
      }
    }
  }
}
.explore-tab-pane ::v-deep {
  .el-tabs__nav-wrap::after,
  .el-tabs__active-bar {
    background-color: rgba(255, 255, 255, 0) !important;
  }
  .el-tabs__item {
    font-size: 13px;
    height: 32px;
    line-height: 32px;
    padding: 0 10px !important;
    margin-right: 6px;
    &.is-active {
      background-color: $color-opacity !important;
      border-radius: 16px;
      font-weight: bold;
    }
  }
  .el-tabs__header {
    margin: 0 !important;
    padding-bottom: 20px !important;
  }
}
</style>
