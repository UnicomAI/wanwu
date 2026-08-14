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
        :range-separator="$t('common.datePicker.at')"
        :start-placeholder="$t('common.datePicker.startDate')"
        :end-placeholder="$t('common.datePicker.endDate')"
        :picker-options="pickerOptions"
        @change="handleChange"
      ></el-date-picker>
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
      // 选区范围（首次选中日期与 6 个月前后边界）
      pickRange: {
        minDate: null,
        maxDate: null,
        lowerLimit: null, // minDate - 6 months
        upperLimit: null, // minDate + 6 months
      },
      pickerOptions: {
        onPick: this.handleOnPick,
        disabledDate: this.handleDisabledDate,
      },
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
      // 切换快捷选项时重置选区状态
      this.resetPickRange();
      this.$emit('handleSetTime', { type: obj[this.radio], time: this.time });
    },
    timestampToDateFormat(timestamp) {
      const dateObj = new Date(timestamp); // 创建Date对象
      const year = dateObj.getFullYear(); // 获取年份
      const month = ('0' + (dateObj.getMonth() + 1)).slice(-2); // 获取月份，并补零
      const day = ('0' + dateObj.getDate()).slice(-2); // 获取日期，并补零
      return `${year}-${month}-${day}`; // 返回转换后的日期格式
    },
    // 监听用户选中的首个日期
    handleOnPick(picked) {
      const { minDate, maxDate } = picked;
      this.pickRange.minDate = minDate;
      this.pickRange.maxDate = maxDate;
      if (minDate && !maxDate) {
        // 首次选完一个日期后，按 365 天作为前后一年的等价天数
        const ONE_DAY_MS = 3600 * 1000 * 24;
        const lower = new Date(minDate.getTime() - 365 * ONE_DAY_MS);
        const upper = new Date(minDate.getTime() + 365 * ONE_DAY_MS);
        // 将时分秒归零，避免边界日期被多算一天
        lower.setHours(0, 0, 0, 0);
        upper.setHours(23, 59, 59, 999);
        this.pickRange.lowerLimit = lower.getTime();
        this.pickRange.upperLimit = upper.getTime();
      }
    },
    // 禁用前后 6 个月之外的日期
    handleDisabledDate(time) {
      const { minDate, maxDate, lowerLimit, upperLimit } = this.pickRange;
      // 已经完整选完一个区间后，重置状态，允许下次重新选
      if (minDate && maxDate) {
        this.resetPickRange();
        return false;
      }
      // 1. 未来日期（明天及以后）一律禁用，今天 0 点 ~ 23:59:59 之间可点
      const todayEnd = new Date();
      todayEnd.setHours(23, 59, 59, 999);
      if (time.getTime() > todayEnd.getTime()) {
        return true;
      }
      // 2. 选完首个日期后，再叠加 ±365 天的窗口限制
      if (minDate && lowerLimit !== null && upperLimit !== null) {
        const t = time.getTime();
        return t < lowerLimit || t > upperLimit;
      }
      return false;
    },
    // date-picker value 变化时同步给父组件
    handleChange(val) {
      if (val && val.length === 2) {
        this.$emit('handleSetTime', { type: obj[this.radio], time: val });
      }
    },
    // 重置选区状态
    resetPickRange() {
      this.pickRange.minDate = null;
      this.pickRange.maxDate = null;
      this.pickRange.lowerLimit = null;
      this.pickRange.upperLimit = null;
    },
  },
  computed: {
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
