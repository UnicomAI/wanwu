<template>
  <div class="statistics_search_time">
    <label>{{$t('common.datePicker.data')}}:</label>
    <div class="search_content">
      <el-radio-group v-model="radio" size="mini" @change="handleRadio">
        <el-radio-button :label="'day'">{{$t('common.datePicker.day')}}</el-radio-button>
        <el-radio-button :label="'week'">{{$t('common.datePicker.week')}}</el-radio-button>
        <el-radio-button :label="'month'">{{$t('common.datePicker.oneMonth')}}</el-radio-button>
      </el-radio-group>
      <el-date-picker
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
        :picker-options="pickerOptions"
        :disabled-date="handleFilterTime"
        @change="handleDateChange"
      >
      </el-date-picker>
      <el-button
        type="primary"
        size="mini"
        :loading="btnLoading"
        @click="handleSearch"
      >
        {{$t('common.button.search')}}
      </el-button>
    </div>
  </div>
</template>
<script>
import {i18n} from "@/lang";

const obj = {
  day: i18n.t('common.datePicker.day'),
  week: i18n.t('common.datePicker.week'),
  month: i18n.t('common.datePicker.oneMonth'),
  cust: i18n.t('common.datePicker.custom'),
};
export default {
  data() {
    const that = this;
    return {
      btnLoading: false,
      radio: "day",
      time: [],
      nowTime: null,
    };
  },
  mounted() {
    // 赋予DefaultValue
    this.time = this.shortcuts;
    // TriggerparentEvent，传递Parameter
    this.$emit("handleSetTime", { type: obj[this.radio], time: this.time });
  },
  methods: {
    handleRadio(val) {
      this.time = this.shortcuts;
      this.radio = val;
      this.$emit("handleSetTime", { type: obj[this.radio], time: this.time });
    },
    handleFilterTime(time) {
      // let time = new Date();
      return time.getTime() > Date.now() - 8.64e7;
    },
    handleDateChange(val) {
      if (val === null) {
        this.time = [];
      }
    },
    timestampToDateFormat(timestamp) {
      const dateObj = new Date(timestamp); // CreateDateObject
      const year = dateObj.getFullYear(); // Get年份
      const month = ("0" + (dateObj.getMonth() + 1)).slice(-2); // Get月份，并补零
      const day = ("0" + dateObj.getDate()).slice(-2); // GetDate，并补零
      return `${year}-${month}-${day}`; // BackConvertafter of DateFormat
    },
    handleSearch() {
      this.$emit("handleSetTime", { type: obj["cust"], time: this.time });
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
              picker.$emit("pick", [start, end]);
            },
          },
          {
            text: this.$t('common.datePicker.week'),
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 6);
              end.setTime(end.getTime());
              picker.$emit("pick", [start, end]);
            },
          },
          {
            text: this.$t('common.datePicker.oneMonth'),
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 29);
              end.setTime(end.getTime());
              picker.$emit("pick", [start, end]);
            },
          },
        ],
        disabledDate(time) {
          if (_this.nowTime) {
            return (
              // time.getTime() > Date.now() - 8.64e7 &&
              // time.getTime() > _this.a.getTime()
              time.getTime() > Date.now() - 8.64e7 ||
              time.getTime() < _this.nowTime.getTime() - 90 * 24 * 3600000 ||
              time.getTime() > _this.nowTime.getTime() + 90 * 24 * 3600000
            );
          } else {
            // return time.getTime() > Date.now() - 8.64e6; //can onlyselect今天And今天previous of Date
            return time.getTime() > Date.now() - 8.64e7; //can onlyselect今天previous of Date，连今天 of Date也不能选
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
      if (this.radio === "day") {
        start.setTime(start.getTime());
        end.setTime(end.getTime());
      } else if (this.radio === "week") {
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
<style lang="sass">
.statistics_search_time
  display: flex
  align-items: center
  position: absolute
  top: 0px
  left: 0
  width: 100%
  min-width: 1200px
  padding: 10px 24px
  z-index: 2001
  .search_content
    margin-left: 10px
    .el-range-editor--mini.el-input__inner
      height: 30px
    .el-button--primary
      margin-left: 10px
    label
      font-size: 14px
    .el-radio-group
      box-shadow: 0 2px 12px 0 rgba(0,0,0,.1)
      border-radius: 28px
      background: #fff
      margin-right: 10px
      padding: 2px
      label
        transform: scale(.89)
        .el-radio-button__inner
          border: 0
          border-radius: 28px

          &:first-child

            .el-radio-button__inner
              border-top-left-radius: 28px
              border-bottom-left-radius: 28px
            &:last-child

              .el-radio-button__inner
                border-top-right-radius: 28px
                border-bottom-right-radius: 28px
        .is-active
          .el-radio-button__inner
            border-radius: 28px
</style>