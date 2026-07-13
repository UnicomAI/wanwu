<template>
  <div class="statistics_common list-common statistics_client_wrapper">
    <div style="padding: 5px 24px">
      <label>{{ $t('statisticsDashboard.appSelect') }}:</label>
      <el-select
        v-model="appParams.module"
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
        v-model="appParams.source"
        :placeholder="$t('statisticsDashboard.sourceFilter')"
        class="no-border-select"
        style="margin-left: 15px"
        clearable
        @change="handleSourceChange"
      >
        <el-option
          v-for="item in sourceOptions"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
      <el-select
        v-if="showSelectAppList.includes(appParams.module)"
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
          <div class="client_dataOverview_content" v-loading="dataLoading">
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
                        ? 'rgb(253,239,241)'
                        : 'rgb(234,251,244)',
                    color: item.des_value < 0 ? '#df1d48' : '#059569',
                  }"
                >
                  <img
                    v-if="item.des_value < 0"
                    src="@/assets/imgs/desc_icon.png"
                    alt=""
                  />
                  <img
                    v-if="item.des_value > 0"
                    src="@/assets/imgs/asc_icon.png"
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
                  echartContent.callResult ? echartContent.callResult.lines : []
                "
                :name="
                  echartContent.callResult
                    ? echartContent.callResult.tableName
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
              <template v-if="module.id === 'byAgent'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByAgent')"
                  dimension="app"
                  :data="rankingData.byAgent || []"
                  :loading="loading"
                />
              </template>
              <template v-else-if="module.id === 'byWorkflow'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByWorkflow')"
                  dimension="app"
                  :data="rankingData.byWorkflow || []"
                  :loading="loading"
                />
              </template>
              <template v-else-if="module.id === 'byChat'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByChat')"
                  dimension="app"
                  :data="rankingData.byChatflow || []"
                  :loading="loading"
                />
              </template>
              <template v-else-if="module.id === 'byRag'">
                <AppRanking
                  :title="$t('statisticsDashboard.appRankingByRag')"
                  dimension="app"
                  :data="rankingData.byRag || []"
                  :loading="loading"
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
  getAppChart,
} from '@/api/statisticsDashboard';
import { AGENT, ShowSelectAppList, TotalTypeObj } from '@/utils/commonSet';

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
      activeTab: 'visual',
      manageDialogVisible: false,
      appTypeObj: TotalTypeObj,
      showSelectAppList: ShowSelectAppList,
      sourceOptions: [
        { label: 'Web', value: 'web' },
        { label: 'OpenAPI', value: 'openapi' },
        { label: 'WebURL', value: 'webURL' },
      ],
      appList: [],
      loading: false,
      dataLoading: false,
      content: {}, // 存储返回的总揽数据
      echartContent: {}, // 存储返回的echart数据
      rankingData: {},
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
          name: this.$t('statisticsDashboard.dailyCallCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'dailyAvgCallCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.dailyCallFailure'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'dailyAvgCallFailure',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.dailyStreamCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'dailyAvgStreamCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.dailyNonStreamCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'dailyAvgNonStreamCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.avgFirstTokenLatency'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgFirstTokenLatency',
          des_value: -9999,
          unit: 'ms',
        },
        {
          name: this.$t('statisticsDashboard.avgCosts'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'avgCosts',
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
          id: 'byAgent',
          name: this.$t('statisticsDashboard.appRankingByAgent'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'byWorkflow',
          name: this.$t('statisticsDashboard.appRankingByWorkflow'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'byChat',
          name: this.$t('statisticsDashboard.appRankingByChat'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'byRag',
          name: this.$t('statisticsDashboard.appRankingByRag'),
          type: 'ranking',
          visible: true,
        },
      ],
      dialogModuleList: [],
      dragIndex: -1,
      appParams: {
        module: AGENT,
        apps: [],
        source: '',
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
    scope: {
      handler() {
        this.fetchApps();
        this.refreshData();
      },
      deep: true,
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
        viewScope: this.scope,
      };
    },
    refreshData() {
      if (!this.timeParams?.time?.length) return;
      const params = this.formatParams({
        ...this.params,
        ...this.appParams,
      });
      this.fetchData(params);
    },
    changeAppType() {
      this.fetchApps();
      this.refreshData();
    },
    handleSourceChange() {
      this.refreshData();
    },
    handleSearch() {
      this.refreshData();
    },
    async fetchApps() {
      this.appList = [];
      this.appParams.apps = [];

      const params = this.formatParams({
        appType: this.appParams.module,
        module: this.appParams.module,
      });
      delete params.apps;
      const res = await getAppSelect(params);
      this.appList = res.data ? res.data.list || [] : [];
    },
    fetchAppList(params) {
      this.dataLoading = true;
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
          this.dataLoading = false;
        });
    },
    async fetchRankingData(params) {
      this.loading = true;
      try {
        const res = await getAppChart(params);
        const { rank, trend } = res.data || {};
        this.rankingData = rank || {};
        this.echartContent = trend || {};
      } catch (err) {
        this.rankingData = {};
      } finally {
        this.loading = false;
      }
    },
    fetchData(params) {
      this.fetchAppList(params);
      this.fetchRankingData(params);
      this.$nextTick(() => {
        this.$refs.appList && this.$refs.appList.getTableData(params);
      });
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
