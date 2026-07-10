<template>
  <div class="statistics_common list-common statistics_client_wrapper">
    <div style="padding: 5px 24px">
      <label>{{ $t('statisticsDashboard.appSelect') }}:</label>
      <el-select
        v-model="appParams.appType"
        :placeholder="$t('statisticsDashboard.appType')"
        class="no-border-select"
        style="margin-left: 15px"
        @change="changeAppType()"
      >
        <el-option
          v-for="key in Object.keys(appTypeObj)"
          :key="key"
          :label="appTypeObj[key]"
          :value="key"
        />
      </el-select>
      <el-select
        v-model="appParams.apps"
        :placeholder="$t('statisticsDashboard.app')"
        class="no-border-select scroll-select"
        style="margin-left: 15px; width: 300px"
        collapse-tags
        clearable
        multiple
        filterable
      >
        <el-option
          v-for="item in appList"
          :key="item.appId"
          :label="item.name"
          :value="item.appId"
        >
          <div class="model-option-content">
            <div class="model-option-content-left">
              <img
                v-if="item?.avatar.path"
                class="model-img"
                :src="convertIcon(item?.avatar.path)"
              />
              <span class="model-name">
                {{ item.name }}
              </span>
            </div>
          </div>
        </el-option>
      </el-select>
      <el-button
        type="primary"
        size="mini"
        :loading="loading"
        style="margin-left: 15px"
        @click="handleSearch"
      >
        {{ $t('common.button.search') }}
      </el-button>
    </div>
    <div class="statistics_content_box scroll-card-container">
      <div class="item_box" style="margin-bottom: 10px">
        <div class="dashboard-tab-header">
          <el-radio-group v-model="activeTab" size="mini">
            <el-radio-button label="visual">
              {{ $t('statisticsDashboard.visualCharts') }}
            </el-radio-button>
            <el-radio-button label="overview">
              {{ $t('statisticsDashboard.statisticsOverview') }}
            </el-radio-button>
          </el-radio-group>
          <el-button
            v-if="activeTab === 'visual'"
            class="manage-btn"
            size="mini"
            plain
            icon="el-icon-setting"
            @click="openManageDialog"
          >
            {{ $t('statisticsDashboard.manage') }}
          </el-button>
        </div>

        <div v-if="activeTab === 'overview'" class="dataOverview">
          <div class="client_dataOverview_content" v-loading="loading">
            <div v-for="(item, index) in count" :key="index" class="card">
              <div class="card-left">
                <div class="card-title">{{ item.name }}</div>
                <div class="card-value">
                  <strong>{{ formatAmount(item.value) }}{{ item.unit }}</strong>
                </div>
              </div>
              <div class="card-right">
                <span
                  v-if="item.des_value !== 0 && item.des_value !== -9999"
                  class="card-tag"
                  :style="{
                    background:
                      item.des_value < 0
                        ? 'rgba(26, 250, 41, 0.1)'
                        : 'rgba(216, 30, 6, 0.1)',
                    color: item.des_value < 0 ? '#1afa29' : '#d81e06',
                  }"
                >
                  <img
                    v-if="item.des_value < 0"
                    src="@/assets/imgs/descend.png"
                    alt=""
                  />
                  <img
                    v-if="item.des_value > 0"
                    src="@/assets/imgs/rise.png"
                    alt=""
                  />
                  {{ (item.des_value > 0 ? '+' : '') + item.des_value + '%' }}
                </span>
                <span v-else class="card-tag default-tag">-</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="activeTab === 'visual'" class="visual-chart-content">
          <div class="chart-modules">
            <div class="data_echart">
              <UserEchart
                :content="
                  echartContent.callTrend ? echartContent.callTrend.lines : []
                "
                :name="
                  echartContent.callTrend
                    ? echartContent.callTrend.tableName
                    : $t('statisticsDashboard.appLineName')
                "
                v-loading="loading"
              ></UserEchart>
            </div>
            <div class="data_echart">
              <UserEchart
                :content="
                  echartContent.failureTrend
                    ? echartContent.failureTrend.lines
                    : []
                "
                :name="
                  echartContent.failureTrend
                    ? echartContent.failureTrend.tableName
                    : $t('statisticsDashboard.appFailureLineName')
                "
                v-loading="loading"
              ></UserEchart>
            </div>
          </div>
          <div class="ranking-modules">
            <div
              v-for="module in visibleModules"
              :key="module.id"
              class="ranking-module"
            >
              <template v-if="module.id === 'rankingByAgent'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByAgent')"
                  dimension="app"
                  :data="getFilteredRankingData(AGENT)"
                  :loading="rankingLoading"
                />
              </template>
              <template v-else-if="module.id === 'rankingByWorkflow'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByWorkflow')"
                  dimension="app"
                  :data="getFilteredRankingData(WORKFLOW)"
                  :loading="rankingLoading"
                />
              </template>
              <template v-else-if="module.id === 'rankingByChat'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByChat')"
                  dimension="app"
                  :data="getFilteredRankingData(CHAT)"
                  :loading="rankingLoading"
                />
              </template>
              <template v-else-if="module.id === 'rankingByRag'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByRag')"
                  dimension="app"
                  :data="getFilteredRankingData(RAG)"
                  :loading="rankingLoading"
                />
              </template>
            </div>
          </div>
        </div>

        <div class="model-list-wrap">
          <AppList
            :params="formatParams({ ...params, ...appParams })"
            ref="appList"
          />
        </div>
      </div>
    </div>

    <el-dialog
      :title="$t('statisticsDashboard.manageConfig')"
      :visible.sync="manageDialogVisible"
      width="500px"
      :close-on-click-modal="false"
      custom-class="model-manage-dialog"
      append-to-body
      @close="closeManageDialog"
    >
      <div class="manage-tips">
        {{ $t('statisticsDashboard.manageConfigTips') }}
      </div>
      <div class="manage-list">
        <div
          v-for="(module, index) in dialogModuleList"
          :key="module.id"
          class="manage-item"
          draggable="true"
          @dragstart="handleDragStart($event, index)"
          @dragover.prevent="handleDragOver($event, index)"
          @drop="handleDrop($event, index)"
          @dragend="handleDragEnd"
        >
          <i class="el-icon-rank drag-icon"></i>
          <span class="manage-item-name">{{ module.name }}</span>
          <el-switch
            v-model="module.visible"
            active-color="#4C7CF6"
            @change="handleModuleVisibleChange"
          ></el-switch>
        </div>
      </div>
      <div slot="footer" class="dialog-footer">
        <el-button size="small" @click="closeManageDialog">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button type="primary" size="small" @click="confirmManageConfig">
          {{ $t('common.button.confirm') }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>
<script>
import UserEchart from '@/components/echart/userEchart.vue';
import AppRanking from './appRanking.vue';
import AppList from './appList.vue';
import { avatarSrc, formatAmount } from '@/utils/util.js';
import {
  getAppData,
  getAppSelect,
  fetchAppList,
} from '@/api/statisticsDashboard';
import { AGENT, AppType, CHAT, WORKFLOW, RAG } from '@/utils/commonSet';

export default {
  components: {
    UserEchart,
    AppRanking,
    AppList,
  },
  props: {
    globalFilterParams: {
      type: Object,
      default: () => ({}),
    },
    timeParams: {
      type: Object,
      default: () => ({ time: [] }),
    },
    scope: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      AGENT,
      WORKFLOW,
      CHAT,
      RAG,
      activeTab: 'visual',
      manageDialogVisible: false,
      appTypeObj: AppType,
      appList: [],
      loading: false,
      rankingLoading: false,
      content: {}, // 存储返回的总揽数据
      echartContent: {}, // 存储返回的echart数据
      rankingData: [],
      count: [
        {
          name: this.$t('statisticsDashboard.appCallCountTotal'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.appCallFailure'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callFailure',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.avgFirstCosts'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgStreamCosts',
          des_value: -9999,
          unit: 'ms',
        },
        {
          name: this.$t('statisticsDashboard.avgCosts'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgNonStreamCosts',
          des_value: -9999,
          unit: 'ms',
        },
        {
          name: this.$t('statisticsDashboard.streamCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'streamCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.nonStreamCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'nonStreamCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
      ],
      moduleList: [
        {
          id: 'rankingByAgent',
          name: this.$t('statisticsDashboard.appRankingByAgent'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'rankingByWorkflow',
          name: this.$t('statisticsDashboard.appRankingByWorkflow'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'rankingByChat',
          name: this.$t('statisticsDashboard.appRankingByChat'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'rankingByRag',
          name: this.$t('statisticsDashboard.appRankingByRag'),
          type: 'ranking',
          visible: true,
        },
      ],
      dialogModuleList: [],
      dragIndex: -1,
      appParams: {
        appType: AGENT,
        apps: [],
      },
    };
  },
  computed: {
    params() {
      return {
        endDate: this.timeParams?.time?.[1],
        startDate: this.timeParams?.time?.[0],
      };
    },
    visibleModules() {
      return this.moduleList.filter(item => item.visible);
    },
  },
  watch: {
    globalFilterParams: {
      handler() {
        this.fetchApps();
        this.refreshData();
      },
      deep: true,
    },
    timeParams: {
      handler(val) {
        if (val?.time?.length) {
          this.refreshData();
        }
      },
      deep: true,
      immediate: true,
    },
  },
  mounted() {
    this.fetchApps();
  },
  methods: {
    formatAmount,
    formatParams(params) {
      return {
        ...params,
        ...this.globalFilterParams,
      };
    },
    refreshData() {
      if (!this.timeParams?.time?.length) return;
      const params = this.formatParams({
        ...this.params,
        ...this.appParams,
      });
      this.fetchData(params);
      this.fetchRankingData(params);
    },
    changeAppType() {
      this.fetchApps();
      this.refreshData();
    },
    handleSearch() {
      this.refreshData();
    },
    async fetchApps() {
      this.appList = [];
      this.appParams.apps = [];

      const res = await getAppSelect(
        this.formatParams({ appType: this.appParams.appType }),
      );
      this.appList = res.data ? res.data.list || [] : [];
    },
    fetchData(params) {
      this.loading = true;
      getAppData(params)
        .then(res => {
          const { overview, trend } = res.data || {};
          this.content = overview || {};
          this.echartContent = trend || {};
          // 解构后台返回的数据，暂存和 count 数组中key对应的数据
          this.count.map(item => {
            item.value = overview[item.key] ? overview[item.key].value : 0;
            item.des_value = overview[item.key]
              ? overview[item.key].periodOverPeriod
              : -9999;
          });
        })
        .finally(() => {
          this.loading = false;
        });
      this.$nextTick(() => {
        if (this.$refs.appList) {
          this.$refs.appList.getTableData(params);
        }
      });
    },
    async fetchRankingData(params) {
      this.rankingLoading = true;
      try {
        // 移除 appType 过滤，一次性获取所有应用类型的数据，前端按类型分组
        const { appType: _ignore, apps: _ignore2, ...rest } = params;
        const res = await fetchAppList({
          ...rest,
          pageNo: 1,
          pageSize: 99999,
        });
        this.rankingData = res?.data?.list || [];
      } catch (err) {
        this.rankingData = [];
      } finally {
        this.rankingLoading = false;
      }
    },
    getFilteredRankingData(appType) {
      return this.rankingData; //this.rankingData.filter(item => item.appType === appType);
    },
    openManageDialog() {
      this.dialogModuleList = JSON.parse(JSON.stringify(this.moduleList));
      this.manageDialogVisible = true;
    },
    closeManageDialog() {
      this.manageDialogVisible = false;
      this.dialogModuleList = [];
      this.dragIndex = -1;
    },
    handleModuleVisibleChange() {
      const visibleCount = this.dialogModuleList.filter(
        item => item.visible,
      ).length;
      if (visibleCount === 0) {
        this.$message.warning(this.$t('statisticsDashboard.atLeastOneModule'));
      }
    },
    handleDragStart(e, index) {
      this.dragIndex = index;
      e.dataTransfer.effectAllowed = 'move';
    },
    handleDragOver(e, index) {
      e.dataTransfer.dropEffect = 'move';
    },
    handleDrop(e, index) {
      if (this.dragIndex === -1 || this.dragIndex === index) return;
      const item = this.dialogModuleList.splice(this.dragIndex, 1)[0];
      this.dialogModuleList.splice(index, 0, item);
      this.dragIndex = index;
    },
    handleDragEnd() {
      this.dragIndex = -1;
    },
    confirmManageConfig() {
      const visibleCount = this.dialogModuleList.filter(
        item => item.visible,
      ).length;
      if (visibleCount === 0) {
        this.$message.error(this.$t('statisticsDashboard.atLeastOneModule'));
        return;
      }
      this.moduleList = JSON.parse(JSON.stringify(this.dialogModuleList));
      this.closeManageDialog();
    },
    convertIcon(iconPath) {
      return iconPath ? avatarSrc(iconPath) : '';
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
@import '@/style/statisticsDashboard.scss';
</style>
