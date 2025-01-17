<template>
  <BarChartGenerator
    :chart-id="chartId"
    :dataset-id-key="chartId"
    :chart-options="chartOptions"
    :chart-data="chartData"
  />
</template>

<script>

import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  TimeScale,
  Title,
  Tooltip,
} from 'chart.js';

import 'chartjs-adapter-moment';

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  LineElement,
  LinearScale,
  CategoryScale,
  PointElement,
  TimeScale,
);

export default {

  name: 'LineChart',

  props: {

    sourceData: Array,

    chartId: String,
  },

  computed: {
    chartData() {
      return {

        labels: (this.sourceData || []).map((row) => new Date(row.date)),

        datasets: [
          {
            label: 'Success',
            borderColor: '#4caf50',
            backgroundColor: '#4caf50',
            data: (this.sourceData || []).map((row) => row.count_by_status.success),
            cubicInterpolationMode: 'monotone',
          },
          {
            label: 'Failed',
            borderColor: '#ff5252',
            backgroundColor: '#ff5252',
            data: (this.sourceData || []).map((row) => row.count_by_status.error),
            cubicInterpolationMode: 'monotone',
          },
          {
            label: 'Stopped',
            borderColor: '#555',
            backgroundColor: '#555',
            data: (this.sourceData || []).map((row) => row.count_by_status.stopped),
            cubicInterpolationMode: 'monotone',
          },
        ],
      };
    },
  },

  data() {
    return {
      chartOptions: {
        scales: {
          x: {
            stacked: true,
            type: 'time',
            time: {
              unit: 'day',
            },
          },
          y: {
            stacked: true,
            ticks: {
              maxTicksLimit: 20,
              stepSize: 1,
            },
          },
        },
        responsive: true,
        maintainAspectRatio: false,
        animation: {
          duration: 0,
        },
      },
    };
  },
};

</script>
