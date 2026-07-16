<template>
  <div class="statistics_common list-common statistics_client_wrapper">
    <div style="padding: 5px 24px">
      <label>{{ $t('statisticsDashboard.apiSelect') }}:</label>
      <el-select
        v-model="apiParams.apiKeyIds"
        :placeholder="$t('statisticsDashboard.apiName')"
        :class="[
          'no-border-select',
          'scroll-select',
          { 'hide-tag-close': isApiSelectedAll },
        ]"
        style="margin-left: 15px; width: 240px"
        multiple
        collapse-tags
        filterable
        clearable
        @change="handleApiNameChange"
      >
        <el-option
          v-for="item in apiNameList"
          :key="item.keyId"
          :label="item.name"
          :value="item.keyId"
        />
      </el-select>
      <el-select
        v-model="apiParams.methodPaths"
        :placeholder="$t('statisticsDashboard.apiPath')"
        class="no-border-select scroll-select"
        style="margin-left: 15px; width: 400px"
        clearable
        multiple
        collapse-tags
        filterable
      >
        <el-option
          v-for="item in apiRoutesList"
          :key="`${item.method}-${item.path}`"
          :label="`${item.method} ${item.path}`"
          :value="`${item.method}-${item.path}`"
        >
          <div class="model-option-content">
            <div class="model-option-content-left">
              <span
                class="model-name"
                :style="`color: ${colorsObj[item.method] || colorsObj['DEFAULT']}`"
              >
                {{ item.method }}
              </span>
              <span class="model-name">
                {{ item.path }}
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
          <!--不上线管理-->
          <!--<el-button
            v-if="activeTab === 'visual'"
            class="manage-btn"
            size="mini"
            plain
            icon="el-icon-setting"
            @click="openManageDialog"
          >
            {{ $t('statisticsDashboard.manage') }}
          </el-button>-->
        </div>

        <div v-if="activeTab === 'overview'" class="dataOverview">
          <div class="client_dataOverview_content" v-loading="dataLoading">
            <div v-for="(item, index) in count" :key="index" class="card">
              <div class="card-left">
                <div class="card-title">{{ item.name }}</div>
                <div class="card-value" v-if="item.unit !== 'ms'">
                  <strong>{{ formatAmount(item.value) }}{{ item.unit }}</strong>
                </div>
                <div class="card-value" v-else>
                  <strong>{{ formatSec(item.value) }}</strong>
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
          <div class="chart-modules" v-if="!rankingVisible">
            <div class="data_echart" v-if="!rankingVisible">
              <UserEchart
                :content="
                  echartContent.apiKeyCalls
                    ? echartContent.apiKeyCalls.lines
                    : []
                "
                :name="
                  echartContent.apiKeyCalls
                    ? echartContent.apiKeyCalls.tableName
                    : $t('statisticsDashboard.apiLineName')
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
                    : $t('statisticsDashboard.apiCallCountChartName')
                "
                v-loading="loading"
              ></UserEchart>
            </div>
          </div>
          <div v-if="rankingVisible" class="chart-modules">
            <div class="data_echart full-width" v-if="rankingVisible">
              <UserEchart
                :content="
                  echartContent.apiKeyCalls
                    ? echartContent.apiKeyCalls.lines
                    : []
                "
                :name="
                  echartContent.apiKeyCalls
                    ? echartContent.apiKeyCalls.tableName
                    : $t('statisticsDashboard.apiLineName')
                "
                v-loading="loading"
              ></UserEchart>
            </div>
          </div>
          <div v-if="rankingVisible" class="chart-modules">
            <ApiRanking
              class="api-ranking-card"
              :title="$t('statisticsDashboard.apiRanking')"
              :data="rankingData?.byApi || []"
              :loading="loading"
            />
            <div
              class="data_echart data_echart_ranking"
              style="width: calc(60% - 20px)"
            >
              <UserEchart
                :content="
                  echartContent.callResult ? echartContent.callResult.lines : []
                "
                :name="
                  echartContent.callResult
                    ? echartContent.callResult.tableName
                    : $t('statisticsDashboard.apiCallCountChartName')
                "
                v-loading="loading"
              ></UserEchart>
            </div>
          </div>
        </div>
        <div class="model-list-wrap">
          <ApiList
            :params="formatParams({ ...params, ...apiParams })"
            ref="apiList"
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
import ApiList from './apiList.vue';
import ApiRanking from './apiRanking.vue';
import { formatAmount, formatSec } from '@/utils/util.js';
import {
  getApiData,
  getApiRoutes,
  getApiSelect,
  getApiChart,
} from '@/api/statisticsDashboard';
import { DEFAULT_APP_ITEM, ALL } from '../../constants';

export default {
  components: {
    UserEchart,
    ApiList,
    ApiRanking,
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
  },
  data() {
    return {
      activeTab: 'visual',
      manageDialogVisible: false,
      apiNameList: [DEFAULT_APP_ITEM],
      apiRoutesList: [],
      colorsObj: {
        GET: '#5CB87A',
        POST: '#E6A23C',
        PATCH: '#A039D3',
        DELETE: '#F56C6C',
        PUT: '#409EFF',
        DEFAULT: '#909399',
      },
      loading: false,
      dataLoading: false,
      echartContent: {}, // 存储返回的echart数据
      rankingData: {},
      count: [
        {
          name: this.$t('statisticsDashboard.callCountTotal'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callCount',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.callFailure'),
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
          id: 'byApi',
          name: this.$t('statisticsDashboard.apiRanking'),
          type: 'ranking',
          visible: true,
        },
      ],
      dialogModuleList: [],
      dragIndex: -1,
      apiParams: {
        apiKeyIds: [ALL],
        methodPaths: [],
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
    isApiSelectedAll() {
      return this.apiParams.apiKeyIds.includes(ALL);
    },
    rankingVisible() {
      return this.moduleList.some(item => item.visible);
    },
  },
  watch: {
    globalFilterParams: {
      handler() {
        this.fetchApiNameList();
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
    this.fetchApiNameList();
    this.fetchApiRoutes();
  },
  methods: {
    formatAmount,
    formatSec,
    formatParams(params) {
      return {
        ...params,
        ...this.globalFilterParams,
      };
    },
    refreshData() {
      if (!this.timeParams?.time?.length) return;
      this.fetchData(this.formatParams({ ...this.params, ...this.apiParams }));
    },
    async fetchApiNameList() {
      this.apiNameList = [DEFAULT_APP_ITEM];
      this.apiParams.apiKeyIds = [ALL];

      const res = await getApiSelect(this.formatParams());
      const list = res.data ? res.data.list || [] : [];
      this.apiNameList = [DEFAULT_APP_ITEM, ...list];
    },
    async fetchApiRoutes() {
      const res = await getApiRoutes();
      this.apiRoutesList = res.data ? res.data.list || [] : [];
    },
    handleApiNameChange(keyIds) {
      if (!keyIds.length) {
        this.apiParams.apiKeyIds = [ALL];
      } else {
        const addKey = keyIds[keyIds.length - 1];
        if (addKey === ALL) {
          this.apiParams.apiKeyIds = [ALL];
        } else {
          const allIndex = this.apiParams.apiKeyIds.findIndex(
            item => item === ALL,
          );
          if (allIndex !== -1) {
            this.apiParams.apiKeyIds.splice(allIndex, 1);
          }
        }
      }
    },
    handleSearch() {
      this.refreshData();
    },
    fetchApiData(params) {
      this.dataLoading = true;
      getApiData(params)
        .then(res => {
          const overview = res.data || {};
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
        const res = await getApiChart(params);
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
      this.fetchApiData(params);
      this.fetchRankingData(params);
      this.$nextTick(() => {
        this.$refs.apiList && this.$refs.apiList.getTableData(params);
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
      this.moduleList = JSON.parse(JSON.stringify(this.dialogModuleList));
      this.closeManageDialog();
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
@import '@/style/statisticsDashboard.scss';
::v-deep .api-ranking-card {
  align-self: flex-start;
}
</style>
