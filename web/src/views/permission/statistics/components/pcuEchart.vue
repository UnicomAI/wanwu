<template>
  <div class="apiEchartBox">
    <span class="title">{{ name }}</span>
    <div ref="api" id="api">
      <el-empty :description="$t('common.noData')"></el-empty>
    </div>
  </div>
</template>
<script>
import * as echarts from "echarts";
import {i18n} from "@/lang";
import { formatAmount } from "@/utils/util.js";

const units = i18n.t("statisticsEcharts.units");

export default {
  props: {
    content: {
      type: Array,
      default() {
        return [];
      },
    },
    name: ''
  },
  data() {
    return {
      api: null,
    };
  },
  watch: {
    content: {
      handler(val) {
        if (val.length > 0) {
          if (this.api) {
            this.api.dispose(); // Destroyprevious of 实例
          }
          this.api = echarts.init(this.$refs.api);

          window.addEventListener("resize", () => {
            this.api.resize();
          });
          this.handleLine();
        }
      },
      deep: true,
    },
  },
  mounted() {
  },
  methods: {
    handleLine() {
      let yData = [];
      let xTime = [];
      const {items = [], lineName} = this.content[0] || {}
      items.map((item) => {
        xTime.push(item.key);
        yData.push(item.value);
      });

      let option = {
        animationDuration: 1000,
        tooltip: {
          trigger: "axis",
          position: "right",
          padding: [5, 8],
          textStyle: {
            color: "#eee",
            fontSize: 13,
          },
          backgroundColor: "rgba(13,5,30,.6)",
          extraCssText: "z-index:1", // 层级
        },
        toolbox: {
          show: true,
          feature: {
            dataView: {
              title: i18n.t('statisticsEcharts.dateView'),
              lang: [i18n.t('statisticsEcharts.dateView'), i18n.t('statisticsEcharts.close'), i18n.t('statisticsEcharts.reload')],
              readOnly: false,
              optionToContent: function (opt) {
                // console.log(opt)
                //该Function可以CustomLististable，optYes给我们提供 of 原始Data of obj。 可print出来DatastructureView
                var axisData = opt.xAxis[0].data; //坐标轴
                var series = opt.series; //折线图 of Data
                var tdHeads = `<td  style="margin-top:10px; padding: 0 15px">${i18n.t('statisticsEcharts.date')}</td>`; //Table头
                var tdBodys = "";
                series.forEach(function (item) {
                  tdHeads += `<td style="padding:5px 15px">${item.name}</td>`;
                });
                var table = `<table border="1" style=";width:100%;margin-left:20px;user-select:text;border-collapse:collapse;font-size:14px;text-align:center"><tbody><tr>${tdHeads} </tr>`;
                for (var i = 0, l = axisData.length; i < l; i++) {
                  for (var j = 0; j < series.length; j++) {
                    if (series[j].data[i] == undefined) {
                      tdBodys += `<td>${"-"}</td>`;
                    } else {
                      tdBodys += `<td>${series[j].data[i]}</td>`;
                    }
                  }
                  table += `<tr><td style="padding: 0 15px">${axisData[i]}</td>${tdBodys}</tr>`;
                  tdBodys = "";
                }
                table += "</tbody></table>";
                return table;
              },
            },
            saveAsImage: {
              title: i18n.t('statisticsEcharts.saveImage')
            },
          },
        },
        legend: {
          show: true,
          data: [lineName],
          x: "center",
          bottom: 10,
          // orient: 'vertical', // 纵向分布
          textStyle: {
            fontSize: 12,
          },
          icon: "rect",
          itemWidth: 20,
          itemHeight: 10,
        },
        grid: {
          top: "10%",
          left: "4%",
          right: "7%",
          bottom: "15%",
          containLabel: true,
        },
        xAxis: {
          type: "category",
          boundaryGap: false,
          data: xTime,
        },
        yAxis: {
          type: "value",
          name:
            this.isLong > 0
              ? `${i18n.t('statisticsEcharts.unit')}(${units[this.isLong]})`
              : units[this.isLong],
          nameLocation: "end",
          nameGap: 15,
          nameTextStyle: {
            padding: [0, 0, 0, -30],
            fontWeight: "bold",
          },
          axisLabel: {
            formatter: (value) => {
              return formatAmount(value, 'object', true).value; // 2位小数
            },
          },
        },
        series: [
          {
            name: lineName,
            data: yData,
            type: "line",
            symbolSize: 5, // 原点size
            smooth: true,
            zlevel: 1, // 层级
            label: {
              position: "right",
              show: false,
              color: "#333",
              fontSize: 13,
              formatter: function (params) {
                return params.data + i18n.t('statisticsEcharts.minute');
              },
            },
            // 折线拐点 of Style
            itemStyle: {
              // 静止 when ：
              color: "#0088FF",
              borderColor: "#0088FF", //拐点 of 边框颜色
              borderWidth: 2,
            },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                {
                  offset: 0,
                  color: "rgba(80,141,255,0.39)",
                },
                {
                  offset: 0.34,
                  color: "rgba(56,155,255,0.05)",
                },
                {
                  offset: 1,
                  color: "rgba(38,197,254,0.00)",
                },
              ]),
            },
            emphasis: {
              focus: "series",
              // 鼠标经  when ：
              itemStyle: {
                  color: "#4CF8C5",
              },
            },
          },
        ],
      };
      this.api.setOption(option);
    },
  },
  computed: {
    isLong() {
      const values = this.content.flatMap(item => item.items);
      if (values.some(item => (item.value || item) >= 100000000)) return 2;
      else if (values.some(item => (item.value || item) >= 100000)) return 1;
      return 0;
    },
  },
  beforeDestroy() {
    if (this.api) {
      this.api.dispose();
    }
  }
};
</script>
<style lang="sass">
@import "@/style/echart.sass"
</style>