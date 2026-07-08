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
        @change="refreshData"
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
      <div class="item_box">
        <div class="dataOverview">
          <span class="title">
            {{ $t('statistics.overview') }}
          </span>
          <div class="client_dataOverview_content" v-loading="loading">
            <div v-for="(item, index) in count" :key="index" class="card">
              <span>
                {{ item.name }}
                <i
                  :style="{
                    background: item.des_value < 0 ? '#1afa29' : '#d81e06',
                  }"
                  :class="{
                    defaultBg: item.des_value === 0 || item.des_value === -9999,
                  }"
                ></i>
              </span>
              <strong>{{ formatAmount(item.value) }}{{ item.unit }}</strong>
              <span>
                {{ item.des }}
                <label
                  :style="{
                    color: item.des_value < 0 ? '#1afa29' : '#d81e06',
                  }"
                  :class="{
                    defaultColor:
                      item.des_value === 0 || item.des_value === -9999,
                  }"
                >
                  {{ item.des_value === -9999 ? '-' : item.des_value + '%' }}
                </label>
                <img
                  v-if="item.des_value < 0 && item.des_value !== -9999"
                  src="@/assets/imgs/descend.png"
                  alt=""
                />
                <img
                  v-if="item.des_value > 0 && item.des_value !== -9999"
                  src="@/assets/imgs/rise.png"
                  alt=""
                />
              </span>
            </div>
          </div>
        </div>
        <div class="data_echart_box">
          <div class="data_echart" style="width: 100%">
            <UserEchart
              :content="
                echartContent.apiCalls ? echartContent.apiCalls.lines : []
              "
              :name="
                echartContent.apiCalls
                  ? echartContent.apiCalls.tableName
                  : $t('statisticsDashboard.apiLineName')
              "
              v-loading="loading"
            ></UserEchart>
          </div>
        </div>
        <div class="dataOverview">
          <span class="title">
            {{ $t('statisticsDashboard.apiList') }}
          </span>
          <div style="margin-top: -20px">
            <ApiList
              :params="formatParams({ ...params, ...apiParams })"
              ref="apiList"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import UserEchart from '@/components/echart/userEchart.vue';
import ApiList from './apiList.vue';
import { formatAmount } from '@/utils/util.js';
import {
  getApiData,
  getApiRoutes,
  getApiSelect,
} from '@/api/statisticsDashboard';
import { DEFAULT_APP_ITEM, ALL } from '../../constants';

export default {
  components: {
    UserEchart,
    ApiList,
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
      content: {}, // 存储返回的总揽数据
      echartContent: {}, // 存储返回的echart数据
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
          name: this.$t('statisticsDashboard.avgStreamCosts'),
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
      this.refreshData();
    },
    handleSearch() {
      this.refreshData();
    },
    fetchData(params) {
      this.loading = true;
      getApiData(params)
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
      this.$refs.apiList.getTableData(params);
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
@import '@/style/statisticsDashboard.scss';
</style>
