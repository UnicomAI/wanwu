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
        </div>
        <div class="dataOverview">
          <span class="title">
            {{ $t('statisticsDashboard.appList') }}
          </span>
          <div style="margin-top: -20px">
            <AppList
              :params="formatParams({ ...params, ...appParams })"
              ref="appList"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import UserEchart from '@/components/echart/userEchart.vue';
import AppList from './appList.vue';
import { avatarSrc, formatAmount } from '@/utils/util.js';
import { getAppData, getAppSelect } from '@/api/statisticsDashboard';
import { AGENT, AppType } from '@/utils/commonSet';

export default {
  components: {
    UserEchart,
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
      appTypeObj: AppType,
      appList: [],
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
      this.fetchData(this.formatParams({ ...this.params, ...this.appParams }));
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
      this.$refs.appList.getTableData(params);
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
