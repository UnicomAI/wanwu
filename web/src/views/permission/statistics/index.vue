<template>
  <div id="statistics_client" class="statistics_common list-common">
    <transition name="el-zoom-in-top">
      <Search
        v-show="searchShow"
        ref="search"
        @handleSetTime="handleSetTime"
      ></Search>
    </transition>
    <div class="statistics_content_box" @scroll="handleScroll">
      <div class="item_box">
        <div class="dataOverview">
          <span class="title">
            {{$t('statistics.overview')}}
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
              <strong>{{ formatAmount(item.value) }}</strong>
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
                  {{
                    item.des_value === -9999 ? "-" : item.des_value + "%"
                  }}
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
          <div class="data_echart">
            <UserEchart
              :content="echartContent.client ? echartContent.client.lines : []"
              :name="echartContent.client ? echartContent.client.tableName : ''"
              v-loading="loading"
            >
            </UserEchart>
          </div>
          <div class="data_echart">
            <PcuEchart
              :content="echartContent.browse ? echartContent.browse.lines : []"
              :name="echartContent.browse ? echartContent.browse.tableName : ''"
              v-loading="loading"
            >
            </PcuEchart>
          </div>
        </div>
      </div>
    </div>
    <el-backtop target=".statistics_content_box"></el-backtop>
  </div>
</template>
<script>
import Search from "./components/search.vue";
import UserEchart from "./components/userEchart.vue";
import PcuEchart from "./components/pcuEchart.vue";
import { formatAmount } from "@/utils/util.js";
import { getData } from "@/api/permission/statistic.js";

export default {
  components: {
    UserEchart,
    PcuEchart,
    Search,
  },
  data() {
    return {
      loading: false,
      content: {}, // 存储Back of 总揽Data
      echartContent: {}, // 存储Back of echartData
      type: "model",
      concurrentUser: {},
      count: [
        {
          name: this.$t('statistics.cumulativeClient'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: "cumulativeClient",
          des_value: -9999,
        },
        {
          name: this.$t('statistics.additionClient'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: "additionClient",
          des_value: -9999,
        },
        {
          name: this.$t('statistics.activeClient'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: "activeClient",
          des_value: -9999,
        },
        {
          name: this.$t('statistics.browse'),
          value: 0,
          des: this.$t('statistics.percentage'),
          key: "browse",
          des_value: -9999,
        },
      ],
      searchShow: true,
      timeout: null, // debounce定 when 
      searchTime: {
        time: [],
      },
    };
  },
  computed: {
    params() {
      return {
        endDate: this.searchTime.time[1],
        startDate: this.searchTime.time[0],
      };
    },
  },
  methods: {
    formatAmount,
    handleScroll(val) {
      this.$refs.search && this.$refs.search.$refs.time.handleClose();
      this.$refs.adminDetail && this.$refs.adminDetail.$refs.sign.handleClose();
      this.$refs.adminDetail &&
      this.$refs.adminDetail.$refs.cycle.handleClose();

      if (val.target.scrollTop >= 50) {
        this.searchShow = false;

        clearTimeout(this.timeout);
        this.timeout = setTimeout(() => {
          document.getElementsByClassName("statistics_search_time")[0].style =
              "background: #fff;box-shadow: 0px 2px 4px 0px rgba(0,0,0,0.1)";
          this.searchShow = true;
        }, 500);
      } else {
        clearTimeout(this.timeout);
        document.getElementsByClassName("statistics_search_time")[0].style =
            "background: transparent;box-shadow: none";
        this.searchShow = true;
      }
    },
    handleSetTime(val) {
      this.loading = true;
      this.searchTime = val;

      const params = {
        startDate: val.time[0],
        endDate: val.time[1],
      }
      getData(params).then((res) => {
        const {overview, trend} = res.data || {}
        this.content = overview || {}
        this.echartContent = trend || {}
        // 解构after台Back of Data，暂存And count Arrayinkey对应 of Data
        this.count.map((item) => {
          item.value = overview[item.key] ? overview[item.key].value : 0;
          item.des_value = overview[item.key] ? overview[item.key].periodOverPeriod : -9999;
        });
      }).finally(() => {
        this.loading = false;
      });
    },
  },
};
</script>
<style lang="scss">
.statistics_common {
  position: relative;
  height: 100%;
  padding: 0;
  padding-top: 48px !important;
  background: #FFF;
  overflow: hidden;
  z-index: 100;
  .statistics_content_box {
    position: relative;
    height: 100%;
    padding: 0 24px 0 24px;
    overflow-y: auto;
    min-width: 1200px;
  }
  .el-radio-button__inner {
    cursor: pointer;
    color: $color;
  }
  .el-radio-button {
    &.is-active {
      span {
        color: #fff !important;
        background: $color !important;
      }
    }
    &.is-disabled {
      span {
        color: #999 !important;
        box-shadow: none;
      }
    }
  }
  .el-backtop {
    i {
      font-size: 20px;
      color: $color;
    }
  }
  .my-pagination {
    /deep/.el-pagination {
      text-align: right;
    }
    .el-pagination.is-background .el-pager li:not(.disabled).active {
      background-color: $color;
      color: #fff;
    }

    .el-pagination .el-select .el-input .el-input__inner {
      padding-right: 25px;
      width: 109px;
      border-color: #cccccc;
    }

    .el-pagination .el-select .el-input .el-input__inner {
      padding-right: 25px;
      width: 109px;
      border-color: #cccccc;
    }

    .el-pager li:hover {
      color: $color;
    }
    .el-pager li.active {
      color: $color;
      cursor: default;
      border: 1px solid $color;
    }

    .el-pagination__editor.el-input .el-input__inner {
      height: 28px;
      background: #ffffff;
      border: 1px solid #cccccc;
    }
    .el-pagination.is-background .el-pager li:not(.disabled):hover {
      color: $color;
    }
  }
  .el-empty {
    position: absolute;
    top:50%;
    left: 50%;
    transform: translate(-50%, -50%);
    padding: 0;
  }
  .el-empty__image {
    width: 15%;
  }
}

#common_header {
  display: flex;
  align-items: center;
  height: 48px;
  padding-left: 24px;
  background: #fff;
  box-shadow: 0 2px 4px 0 rgba(0, 0, 0, .1);
  position: fixed;
  top: 0;
  width: 100%;
  z-index: 2002;
  span {
    font-size: 14px;
  }
}

#statistics_client {
  .item_box {
    .client_const {
      display: flex;
      padding: 20px 0;
      justify-content: space-between;
      background: #fff;
      margin-bottom: 20px;
      border-radius: 5px;

      span {
        display: flex;
        justify-content: center;
        align-items: center;
        width: calc(100% / 3);
        height: 100px;
        border-left: 1px solid #e8e9eb;

        &:first-child {
          border: 0;
        }

        img {
          height: 70px;
          margin-right: 20px;
        }

        div {
          display: flex;
          flex-direction: column;
          justify-content: space-between;
          height: 70px;
          font-size: 15px;

          strong {
            font-size: 20px;
          }
        }
      }
    }

    .defaultColor {
      color: #abb0b5 !important;
    }

    .defaultBg {
      background: #abb0b5 !important;
    }

    .data_echart_box {
      display: flex;
      justify-content: space-between;

      .data_echart {
        width: calc(50% - 10px);
      }
    }

    .data_echart {
      display: inline-block;
      width: 100%;
      position: relative;
      margin-bottom: 20px;
      background: #fff;
      border-radius: 5px;

      .title {
        display: block;
        margin-bottom: 20px;
      }

      .el-radio-group {
        margin-bottom: 20px;
      }
    }

    .dataOverview {
      padding: 20px;
      margin-bottom: 20px;
      background: #fff;
      flex: 1;
      border-radius: 5px;

      .client_dataOverview_content {
        display: flex;
        justify-content: space-around;
        margin-top: 20px;

        .card {
          position: relative;
          flex: 1;
          min-width: 120px;
          height: 120px;
          background: rgb(245, 246, 249);
          border-radius: 4px;
          padding: 15px 0;
          margin-bottom: 20px;
          margin-right: 10px;
          margin-left: 10px;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: space-between;
          flex-shrink: 0;
          box-sizing: border-box;

          span {
            position: relative;

            &:last-child {
              display: flex;
              align-items: center;
              justify-content: space-between;
              font-size: 12px;
              color: rgb(171, 176, 181);
            }

            label {
              color: #303133;
              margin-left: 10px;
              font-weight: bold;
              font-size: 14px;
            }

            img {
              width: 13px;
              vertical-align: middle;
            }
          }

          strong {
            font-size: 15px;
          }

          i {
            position: absolute;
            width: 8px;
            height: 8px;
            border-radius: 50%;
            top: 5px;
            left: -13px;
            z-index: 1;
          }
        }
      }
    }

    .title {
      position: relative;
      font-size: 14px;
      font-weight: bold;
      padding-left: 10px;

      &::after {
        content: "";
        width: 3px;
        height: 15px;
        background: $color;
        position: absolute;
        left: 0;
        top: 50%;
        transform: translate(0, -50%);
      }

      label {
        font-size: 10px;
        color: rgb(171, 176, 181);
      }
    }
  }
}
</style>
