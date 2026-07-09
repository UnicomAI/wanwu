<template>
  <div class="statistics_common list-common statistics_client_wrapper">
    <div style="padding: 5px 24px">
      <label>{{ $t('statisticsDashboard.modelSelect') }}:</label>
      <el-select
        v-model="modelParams.modelType"
        :placeholder="$t('modelAccess.table.modelType')"
        class="no-border-select"
        style="margin-left: 15px"
        @change="changeModelType()"
      >
        <el-option
          v-for="item in modelTypeList"
          :key="item.key"
          :label="item.name"
          :value="item.key"
        />
      </el-select>
      <el-select
        v-model="modelParams.models"
        :placeholder="$t('statisticsDashboard.model')"
        class="no-border-select scroll-select"
        style="margin-left: 15px; width: 300px"
        collapse-tags
        clearable
        multiple
        filterable
      >
        <el-option
          v-for="item in modelList"
          v-if="item.displayName"
          :key="item.modelId"
          :label="item.displayName"
          :value="item.modelId"
        >
          <div class="model-option-content">
            <div class="model-option-content-left">
              <img
                class="model-img"
                :src="convertModelIcon(item?.avatar.path)"
              />
              <span class="model-name">
                {{ item.displayName }}
              </span>
            </div>

            <div
              class="model-select-tags"
              v-if="item.tags && item.tags.length > 0"
            >
              <span
                v-for="(tag, tagIdx) in item.tags"
                :key="tagIdx"
                class="model-select-tag"
              >
                {{ tag.text }}
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
                  echartContent.tokensUsage
                    ? echartContent.tokensUsage.lines
                    : []
                "
                :name="
                  echartContent.tokensUsage
                    ? echartContent.tokensUsage.tableName
                    : ''
                "
                v-loading="loading"
              ></UserEchart>
            </div>
            <div class="data_echart">
              <UserEchart
                :content="
                  echartContent.modelCalls ? echartContent.modelCalls.lines : []
                "
                :name="
                  echartContent.modelCalls
                    ? echartContent.modelCalls.tableName
                    : ''
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
              <template v-if="module.id === 'rankingByModel'">
                <ModelRanking
                  :title="$t('statisticsDashboard.modelRankingByModel')"
                  dimension="model"
                  :data="rankingData"
                  :loading="rankingLoading"
                  :model-map="modelMap"
                />
              </template>
              <template v-else-if="module.id === 'rankingByUser'">
                <ModelRanking
                  :title="$t('statisticsDashboard.modelRankingByUser')"
                  dimension="user"
                  :data="rankingData"
                  :loading="rankingLoading"
                  :model-map="modelMap"
                />
              </template>
              <template v-else-if="module.id === 'rankingByOrg'">
                <ModelRanking
                  :title="$t('statisticsDashboard.modelRankingByOrg')"
                  dimension="org"
                  :data="rankingData"
                  :loading="rankingLoading"
                  :model-map="modelMap"
                />
              </template>
            </div>
          </div>
        </div>

        <div class="model-list-wrap">
          <ModelList
            :params="formatParams({ ...params, ...modelParams })"
            :model-map="modelMap"
            ref="modelList"
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
import ModelRanking from './modelRanking.vue';
import ModelList from './modelList.vue';
import { avatarSrc, formatAmount, getModelDefaultIcon } from '@/utils/util.js';
import {
  getModelData,
  getModelSelect,
  fetchModelList,
} from '@/api/statisticsDashboard';
import { MODEL_TYPE, LLM } from '@/views/modelAccess/constants';

export default {
  components: {
    UserEchart,
    ModelRanking,
    ModelList,
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
      activeTab: 'visual',
      manageDialogVisible: false,
      modelTypeList: MODEL_TYPE,
      modelList: [],
      loading: false,
      rankingLoading: false,
      content: {},
      echartContent: {},
      rankingData: [],
      count: [
        {
          name: this.$t('statisticsDashboard.tokenTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'totalTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
        },
        {
          name: this.$t('statisticsDashboard.promptTokensTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'promptTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
        },
        {
          name: this.$t('statisticsDashboard.completionTokensTotals'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'completionTokensTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.quantity'),
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
          name: this.$t('statisticsDashboard.callCount'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callCountTotal',
          des_value: -9999,
          unit: this.$t('statisticsDashboard.frequency'),
        },
        {
          name: this.$t('statisticsDashboard.callFailure'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: 'callFailureTotal',
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
      ],
      moduleList: [
        {
          id: 'rankingByModel',
          name: this.$t('statisticsDashboard.modelRankingByModel'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'rankingByUser',
          name: this.$t('statisticsDashboard.modelRankingByUser'),
          type: 'ranking',
          visible: true,
        },
        {
          id: 'rankingByOrg',
          name: this.$t('statisticsDashboard.modelRankingByOrg'),
          type: 'ranking',
          visible: true,
        },
      ],
      dialogModuleList: [],
      dragIndex: -1,
      modelParams: {
        modelType: LLM,
        models: [],
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
    modelMap() {
      const map = {};
      this.modelList.forEach(item => {
        if (item.modelId) {
          map[item.modelId] = item;
        }
      });
      return map;
    },
  },
  watch: {
    globalFilterParams: {
      handler() {
        this.fetchModels();
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
    this.fetchModels();
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
        ...this.modelParams,
      });
      this.fetchData(params);
      this.fetchRankingData(params);
    },
    changeModelType() {
      this.fetchModels();
      this.refreshData();
    },
    handleSearch() {
      this.refreshData();
    },
    async fetchModels() {
      this.modelList = [];
      this.modelParams.models = [];

      const res = await getModelSelect(
        this.formatParams({ modelType: this.modelParams.modelType }),
      );
      this.modelList = res.data ? res.data.list || [] : [];
    },
    fetchData(params) {
      this.loading = true;
      getModelData(params)
        .then(res => {
          const { overview, trend } = res.data || {};
          this.content = overview || {};
          this.echartContent = trend || {};
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
        if (this.$refs.modelList) {
          this.$refs.modelList.getTableData(params);
        }
      });
    },
    async fetchRankingData(params) {
      this.rankingLoading = true;
      try {
        const res = await fetchModelList({
          ...params,
          pageNo: 1,
          pageSize: 99999,
        });
        this.rankingData = res.data?.list || [];
      } finally {
        this.rankingLoading = false;
      }
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
    convertModelIcon(iconPath) {
      return iconPath ? avatarSrc(iconPath) : getModelDefaultIcon();
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
@import '@/style/statisticsDashboard.scss';
</style>
