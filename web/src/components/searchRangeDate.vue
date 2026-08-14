<template>
  <div class="statistics_search_time">
    <div class="search_content">
      <el-radio-group v-model="radio" size="mini" @change="handleRadio">
        <!--<el-radio-button :label="'day'">
          {{ $t('common.datePicker.day') }}
        </el-radio-button>-->
        <el-radio-button :label="'week'">
          {{ $t('common.datePicker.week') }}
        </el-radio-button>
        <el-radio-button :label="'month'">
          {{ $t('common.datePicker.oneMonth') }}
        </el-radio-button>
      </el-radio-group>
      <el-date-picker
        ref="time"
        v-model="time"
        size="mini"
        :clearable="false"
        type="daterange"
        value-format="yyyy-MM-dd"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
      ></el-date-picker>
      <!--<el-date-picker
        ref="time"
        size="mini"
        v-model="time"
        type="daterange"
        :clearable="false"
        align="right"
        value-format="yyyy-MM-dd"
        unlink-panels
        :range-separator="$t('common.datePicker.at')"
        :start-placeholder="$t('common.datePicker.startDate')"
        :end-placeholder="$t('common.datePicker.endDate')"
        :disabled-date="handleFilterTime"
        @change="handleDateChange"
      ></el-date-picker>-->
    </div>
  </div>
</template>
<script>
import { i18n } from '@/lang';

const obj = {
  day: i18n.t('common.datePicker.day'),
  week: i18n.t('common.datePicker.week'),
  month: i18n.t('common.datePicker.oneMonth'),
  cust: i18n.t('common.datePicker.custom'),
};

export default {
  data() {
    return {
      radio: 'week',
      time: [],
      nowTime: null,
    };
  },
  mounted() {
    // 赋予默认值
    this.time = this.shortcuts;
    // 触发父级事件，传递参数
    this.$emit('handleSetTime', { type: obj[this.radio], time: this.time });
  },
  methods: {
    handleRadio(val) {
      this.time = this.shortcuts;
      this.radio = val;
      this.$emit('handleSetTime', { type: obj[this.radio], time: this.time });
    },
    handleFilterTime(time) {
      // let time = new Date();
      return time.getTime() > Date.now() - 8.64e7;
    },
    handleDateChange(val) {
      if (val === null) {
        this.time = [];
        return;
      }
      this.$emit('handleSetTime', { type: obj['cust'], time: val });
    },
    timestampToDateFormat(timestamp) {
      const dateObj = new Date(timestamp); // 创建Date对象
      const year = dateObj.getFullYear(); // 获取年份
      const month = ('0' + (dateObj.getMonth() + 1)).slice(-2); // 获取月份，并补零
      const day = ('0' + dateObj.getDate()).slice(-2); // 获取日期，并补零
      return `${year}-${month}-${day}`; // 返回转换后的日期格式
    },
  },
  computed: {
    pickerOptions() {
      const _this = this;
      return {
        shortcuts: [
          {
            text: this.$t('common.datePicker.day'),
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime());
              end.setTime(end.getTime());
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: this.$t('common.datePicker.week'),
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 6);
              end.setTime(end.getTime());
              picker.$emit('pick', [start, end]);
            },
          },
          {
            text: this.$t('common.datePicker.oneMonth'),
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 29);
              end.setTime(end.getTime());
              picker.$emit('pick', [start, end]);
            },
          },
        ],
        disabledDate(time) {
          if (_this.nowTime) {
            return (
              time.getTime() > Date.now() ||
              time.getTime() < _this.nowTime.getTime() - 90 * 24 * 3600000 ||
              time.getTime() > _this.nowTime.getTime() + 90 * 24 * 3600000
            );
          } else {
            return time.getTime() > Date.now(); // 只能选择今天及今天之前的日期
          }
        },
        onPick(picker, date, dateString) {
          _this.nowTime = picker.minDate;
        },
      };
    },
    shortcuts() {
      const end = new Date();
      const start = new Date();
      if (this.radio === 'day') {
        start.setTime(start.getTime());
        end.setTime(end.getTime());
      } else if (this.radio === 'week') {
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 6);
        end.setTime(end.getTime());
      } else {
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 29);
        end.setTime(end.getTime());
      }
      return [
        this.timestampToDateFormat(start),
        this.timestampToDateFormat(end),
      ];
    },
  },
};
</script>
<style lang="scss">
.statistics_search_time {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 10px 24px;
  z-index: 2001;

  .search_content {
    margin-left: 10px;

    .el-range-editor--mini.el-input__inner {
      height: 30px;
      box-shadow:
        0 0 15px 0 rgba(89, 104, 178, 0.06),
        0 15px 20px 0 rgba(89, 104, 178, 0.06);
      border: none;
    }

    .el-button--primary {
      margin-left: 10px;
    }

    label {
      font-size: 14px;
    }

    .el-radio-group {
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
      border-radius: 28px;
      background: #fff;
      margin-right: 10px;
      padding: 2px;

      label {
        transform: scale(0.89);

        .el-radio-button__inner {
          border: 0;
          border-radius: 28px;
        }

        &:first-child {
          .el-radio-button__inner {
            border-top-left-radius: 28px;
            border-bottom-left-radius: 28px;
          }
        }

        &:last-child {
          .el-radio-button__inner {
            border-top-right-radius: 28px;
            border-bottom-right-radius: 28px;
          }
        }

        &.is-active {
          .el-radio-button__inner {
            border-radius: 28px;
          }
        }
      }
    }
  }
}
</style>
