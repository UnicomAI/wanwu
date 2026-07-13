<template>
  <div class="apiEchartBox">
    <span class="title">{{ name }}</span>
    <div ref="api" id="api">
      <el-empty v-if="!hasData" :description="$t('common.noData')"></el-empty>
    </div>
  </div>
</template>

<script>
import * as echarts from 'echarts';
import { i18n } from '@/lang';
import { formatAmount } from '@/utils/util';

const units = i18n.t('statisticsEcharts.units');

export default {
  props: {
    content: {
      type: Array,
      default() {
        return [];
      },
    },
    name: '',
  },
  data() {
    return {
      api: null,
      resizeHandler: null,
      roHandler: null,
    };
  },
  computed: {
    hasData() {
      return (
        this.content &&
        this.content.length > 0 &&
        this.content.some(item => item.items && item.items.length > 0)
      );
    },
    isLong() {
      const values = this.content.flatMap(item => item.items || []);
      if (values.some(item => (item.value || item) >= 100000000)) return 2;
      else if (values.some(item => (item.value || item) >= 100000)) return 1;
      return 0;
    },
  },
  watch: {
    content: {
      handler(val) {
        this.renderChart();
      },
      deep: true,
    },
  },
  mounted() {
    this.renderChart();
  },
  methods: {
    renderChart() {
      const val = this.content;
      if (!val || val.length === 0 || !this.$refs.api) {
        return;
      }
      if (this.api) {
        this.api.dispose();
      }
      this.api = echarts.init(this.$refs.api);
      if (!this.resizeHandler) {
        this.resizeHandler = () => {
          if (this.api) {
            this.api.resize();
          }
        };
        window.addEventListener('resize', this.resizeHandler);
      }
      if (!this.roHandler && typeof ResizeObserver !== 'undefined') {
        this._roInitialized = false;
        this.roHandler = () => {
          // 跳过首次回调，避免打断 setOption 的初始动效
          if (!this._roInitialized) {
            this._roInitialized = true;
            return;
          }
          if (this.api) {
            this.api.resize();
          }
        };
        const ro = new ResizeObserver(this.roHandler);
        ro.observe(this.$refs.api);
        this._ro = ro;
      }
      this.handleLine();
    },
    handleLine() {
      let seriesData = [];
      let legendData = [];
      let xTime = [];

      // 获取x轴数据
      if (this.content.length > 0 && this.content[0].items) {
        xTime = this.content[0].items.map(item => item.key);
      }

      // 构建每个系列的数据
      this.content.forEach((line, index) => {
        legendData.push(line.lineName);

        let yData = [];
        if (line.items) {
          yData = line.items.map(item => item.value);
        }

        seriesData.push({
          name: line.lineName,
          data: yData,
          type: 'line',
          symbolSize: 5,
          smooth: true,
          zlevel: 1,
          label: {
            position: 'right',
            show: false,
            color: '#333',
            fontSize: 13,
            formatter: function (params) {
              return params.data + i18n.t('statisticsEcharts.minute');
            },
          },
          itemStyle: {
            color: this.getColorByIndex(index),
            borderColor: this.getColorByIndex(index),
            borderWidth: 2,
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              {
                offset: 0,
                color: this.getAreaColorByIndex(index, 0.39),
              },
              {
                offset: 0.34,
                color: this.getAreaColorByIndex(index, 0.05),
              },
              {
                offset: 1,
                color: this.getAreaColorByIndex(index, 0.0),
              },
            ]),
          },
          emphasis: {
            focus: 'series',
            itemStyle: {
              color: '#4CF8C5',
            },
          },
        });
      });

      let option = {
        animationDuration: 1000,
        tooltip: {
          trigger: 'axis',
          position: 'right',
          padding: [5, 8],
          textStyle: {
            color: '#eee',
            fontSize: 13,
          },
          backgroundColor: 'rgba(13,5,30,.6)',
          extraCssText: 'z-index:1',
        },
        toolbox: {
          show: true,
          feature: {
            dataView: {
              title: i18n.t('statisticsEcharts.dateView'),
              lang: [
                i18n.t('statisticsEcharts.dateView'),
                i18n.t('statisticsEcharts.close'),
              ],
              readOnly: true,
              optionToContent: function (opt) {
                var axisData = opt.xAxis[0].data;
                var series = opt.series;
                var tdHeads = `<td  style="margin-top:10px; padding: 0 15px">${i18n.t('statisticsEcharts.date')}</td>`;
                var tdBodys = '';
                series.forEach(function (item) {
                  tdHeads += `<td style="padding:5px 15px">${item.name}</td>`;
                });
                var table = `<table border="1" style=";width:100%;margin-left:20px;user-select:text;border-collapse:collapse;font-size:14px;text-align:center"><tbody><tr>${tdHeads} </tr>`;
                for (var i = 0, l = axisData.length; i < l; i++) {
                  for (var j = 0; j < series.length; j++) {
                    if (series[j].data[i] == undefined) {
                      tdBodys += `<td>${'-'}</td>`;
                    } else {
                      tdBodys += `<td>${series[j].data[i]}</td>`;
                    }
                  }
                  table += `<tr><td style="padding: 0 15px">${axisData[i]}</td>${tdBodys}</tr>`;
                  tdBodys = '';
                }
                table += '</tbody></table>';
                return table;
              },
            },
            saveAsImage: {
              title: i18n.t('statisticsEcharts.saveImage'),
            },
          },
        },
        legend: {
          show: true,
          data: legendData,
          x: 'center',
          bottom: 10,
          textStyle: {
            fontSize: 12,
          },
          icon: 'rect',
          itemWidth: 20,
          itemHeight: 10,
        },
        grid: {
          top: '10%',
          left: '4%',
          right: '7%',
          bottom: '15%',
          containLabel: true,
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: xTime,
        },
        yAxis: {
          type: 'value',
          name:
            this.isLong > 0
              ? `${i18n.t('statisticsEcharts.unit')}(${units[this.isLong]})`
              : units[this.isLong],
          nameLocation: 'end',
          nameGap: 15,
          nameTextStyle: {
            padding: [0, 0, 0, -30],
            fontWeight: 'bold',
          },
          axisLabel: {
            formatter: value => {
              return formatAmount(value, 'object', true).value;
            },
          },
        },
        series: seriesData,
      };

      this.api.setOption(option);
    },

    // 根据索引获取不同颜色
    getColorByIndex(index) {
      const colors = ['#0088FF', '#FF9F40', '#1DD1A1', '#FF6B6B', '#5F27CD'];
      return colors[index % colors.length];
    },

    // 根据索引获取区域颜色
    getAreaColorByIndex(index, opacity) {
      const baseColors = [
        '80,141,255',
        '255,159,64',
        '29,209,161',
        '255,107,107',
        '95,39,205',
      ];
      const baseColor = baseColors[index % baseColors.length];
      return `rgba(${baseColor},${opacity})`;
    },
  },
  beforeDestroy() {
    if (this.api) {
      this.api.dispose();
    }
    if (this.resizeHandler) {
      window.removeEventListener('resize', this.resizeHandler);
    }
    if (this._ro) {
      this._ro.disconnect();
    }
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/echart.scss';
</style>
