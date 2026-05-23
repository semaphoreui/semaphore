<template>
  <LineChartGenerator
    :chart-id="chartId"
    :dataset-id-key="chartId"
    :chart-options="chartOptions"
    :chart-data="chartData"
    style="height: 220px;"
  />
</template>

<script>

import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  TimeScale,
  Title,
  Tooltip,
} from 'chart.js';

import './chartjs-adapter-day';

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  Filler,
  LineElement,
  LinearScale,
  CategoryScale,
  PointElement,
  TimeScale,
);

function formatBytes(bytes) {
  if (bytes == null || Number.isNaN(bytes)) {
    return '—';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = bytes;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(v >= 100 || i === 0 ? 0 : 1)} ${units[i]}`;
}

export default {

  name: 'RedisMemoryChart',

  props: {
    history: {
      type: Array,
      default: () => [],
    },
    chartId: {
      type: String,
      default: 'redis-memory-chart',
    },
    windowMs: {
      type: Number,
      default: 60 * 60 * 1000,
    },
    nowTs: {
      type: Number,
      default: () => Date.now(),
    },
  },

  computed: {
    chartData() {
      return {
        datasets: [
          {
            label: 'Used memory',
            borderColor: '#1976d2',
            backgroundColor: 'rgba(25, 118, 210, 0.15)',
            data: (this.history || []).map((row) => ({
              x: row.time,
              y: row.bytes,
            })),
            cubicInterpolationMode: 'monotone',
            fill: true,
            pointRadius: 0,
            borderWidth: 2,
          },
        ],
      };
    },

    chartOptions() {
      const end = this.nowTs;
      const start = end - this.windowMs;
      return {
        scales: {
          x: {
            type: 'time',
            min: start,
            max: end,
            time: {
              unit: 'minute',
              tooltipFormat: 'HH:mm:ss',
              displayFormats: { minute: 'HH:mm' },
            },
            ticks: { maxRotation: 0, autoSkip: true, maxTicksLimit: 8 },
          },
          y: {
            beginAtZero: true,
            ticks: {
              callback: (v) => formatBytes(v),
            },
          },
        },
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => formatBytes(ctx.parsed.y),
            },
          },
        },
        responsive: true,
        maintainAspectRatio: false,
        animation: { duration: 0 },
      };
    },
  },
};

</script>
