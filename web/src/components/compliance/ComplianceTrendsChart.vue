<template>
  <div class="compliance-chart">
    <v-container v-if="!data || data.length === 0" class="text-center py-8">
      <v-icon size="48" color="grey">mdi-chart-line</v-icon>
      <div class="text-h6 grey--text mt-2">No data available</div>
    </v-container>

    <div v-else class="compliance-chart__container" ref="chartContainer"></div>

    <!-- Chart Legend -->
    <div v-if="showLegend && data && data.length > 0" class="compliance-chart__legend">
      <div class="compliance-chart__legend-item">
        <div
          class="compliance-chart__legend-item__color"
          :style="{ backgroundColor: color }"
        ></div>
        <span>{{ legendLabel }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import * as echarts from 'echarts';

export default {
  name: 'ComplianceTrendsChart',

  props: {
    data: {
      type: Array,
      default: () => [],
    },
    title: {
      type: String,
      default: 'Trend',
    },
    color: {
      type: String,
      default: '#1976d2',
    },
    chartType: {
      type: String,
      default: 'line',
      validator: (value) => ['line', 'bar', 'area', 'pie'].includes(value),
    },
    showLegend: {
      type: Boolean,
      default: true,
    },
    legendLabel: {
      type: String,
      default: 'Value',
    },
    height: {
      type: Number,
      default: 300,
    },
    showTooltip: {
      type: Boolean,
      default: true,
    },
    showGrid: {
      type: Boolean,
      default: true,
    },
    smooth: {
      type: Boolean,
      default: true,
    }
  },

  data() {
    return {
      chart: null,
    };
  },

  watch: {
    data: {
      handler() {
        this.updateChart();
      },
      deep: true,
    },
  },

  mounted() {
    this.initChart();
  },

  beforeDestroy() {
    if (this.chart) {
      this.chart.dispose();
    }
  },

  methods: {
    initChart() {
      if (this.$refs.chartContainer) {
        this.chart = echarts.init(this.$refs.chartContainer);
        this.updateChart();

        // Handle window resize
        window.addEventListener('resize', this.handleResize);
      }
    },

    updateChart() {
      if (!this.chart || !this.data || this.data.length === 0) {
        return;
      }

      const dates = this.data.map((item) => {
        const date = new Date(item.date);
        return date.toLocaleDateString();
      });

      const values = this.data.map((item) => item.value);

      const option = {
        title: {
          text: this.title,
          left: 'center',
          textStyle: {
            fontSize: 14,
            fontWeight: 'normal',
          },
        },
        tooltip: this.showTooltip ? {
          trigger: 'axis',
          formatter: (params) => {
            if (params && params.length > 0) {
              const data = params[0];
              const originalData = this.data[data.dataIndex];
              return `<div><strong>${data.name}</strong><br/>Value: ${data.value}<br/>Count: ${originalData.count || 0}</div>`;
            }
            return '';
          },
        } : { show: false },
        grid: this.showGrid ? {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        } : { show: false },
        xAxis: this.chartType !== 'pie' ? {
          type: 'category',
          data: dates,
          axisLabel: {
            rotate: 45,
            fontSize: 10,
          },
        } : undefined,
        yAxis: this.chartType !== 'pie' ? {
          type: 'value',
          axisLabel: {
            formatter: (value) => {
              if (value >= 1000) {
                return (value / 1000).toFixed(1) + 'k';
              }
              return value.toString();
            },
          },
        } : undefined,
        series: this.getSeriesConfig(values),
      };

      this.chart.setOption(option);
    },

    getSeriesConfig(values) {
      const baseConfig = {
        name: this.title,
        data: values,
        itemStyle: {
          color: this.color,
        },
      };

      switch (this.chartType) {
        case 'bar':
          return [{
            ...baseConfig,
            type: 'bar',
            barWidth: '60%',
          }];

        case 'area':
          return [{
            ...baseConfig,
            type: 'line',
            smooth: this.smooth,
            lineStyle: {
              color: this.color,
              width: 2,
            },
            areaStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  {
                    offset: 0,
                    color: this.color + '40',
                  },
                  {
                    offset: 1,
                    color: this.color + '10',
                  },
                ],
              },
            },
          }];

        case 'pie':
          return [{
            ...baseConfig,
            type: 'pie',
            radius: ['40%', '70%'],
            data: this.data.map((item) => ({
              name: new Date(item.date).toLocaleDateString(),
              value: item.value,
            })),
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)',
              },
            },
          }];

        case 'line':
        default:
          return [{
            ...baseConfig,
            type: 'line',
            smooth: this.smooth,
            lineStyle: {
              color: this.color,
              width: 2,
            },
            markPoint: {
              data: [
                {
                  type: 'max',
                  name: 'Max',
                  symbol: 'pin',
                  symbolSize: 50,
                  label: {
                    show: true,
                    formatter: 'Max',
                    position: 'top',
                  },
                },
              ],
            },
          }];
      }
    },

    handleResize() {
      if (this.chart) {
        this.chart.resize();
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.compliance-chart {
  &__container {
    position: relative;
    height: 300px;
    width: 100%;
  }

  &__legend {
    display: flex;
    justify-content: center;
    margin-top: 16px;
    gap: 16px;

    &-item {
      display: flex;
      align-items: center;
      font-size: 0.875rem;

      &__color {
        width: 12px;
        height: 12px;
        border-radius: 50%;
        margin-right: 8px;
      }
    }
  }
}

// Responsive design
@media (max-width: 768px) {
  .compliance-chart {
    &__container {
      height: 250px;
    }
  }
}

@media (max-width: 480px) {
  .compliance-chart {
    &__container {
      height: 200px;
    }
  }
}
</style>