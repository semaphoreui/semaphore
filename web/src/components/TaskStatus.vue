<template>
  <v-chip v-if="status" style="font-weight: bold;" :color="getStatusColor(status)">
    <v-icon v-if="status !== 'running'" left>{{ getStatusIcon(status) }}</v-icon>
    <IndeterminateProgressCircular v-else style="margin-left: -5px;" />
    {{ humanizeStatus(status) }}
  </v-chip>
</template>
<script>
import IndeterminateProgressCircular from '@/components/IndeterminateProgressCircular.vue';

const TaskStatus = Object.freeze({
  WAITING: 'waiting',
  STARTING: 'starting',
  WAITING_CONFIRMATION: 'waiting_confirmation',
  CONFIRMED: 'confirmed',
  RUNNING: 'running',
  SUCCESS: 'success',
  ERROR: 'error',
  STOPPING: 'stopping',
  STOPPED: 'stopped',
});

export default {
  components: { IndeterminateProgressCircular },
  props: {
    status: String,
  },

  methods: {
    getStatusIcon(status) {
      switch (status) {
        case TaskStatus.WAITING:
          return 'mdi-alarm';
        case TaskStatus.STARTING:
          return 'mdi-play-circle';
        case TaskStatus.RUNNING:
          return '';
        case TaskStatus.SUCCESS:
          return 'mdi-check-circle';
        case TaskStatus.ERROR:
          return 'mdi-information';
        case TaskStatus.STOPPING:
          return 'mdi-stop-circle';
        case TaskStatus.STOPPED:
          return 'mdi-stop-circle';
        case TaskStatus.CONFIRMED:
          return 'mdi-check-circle';
        case TaskStatus.WAITING_CONFIRMATION:
          return 'mdi-pause-circle';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },

    humanizeStatus(status) {
      switch (status) {
        case TaskStatus.WAITING:
          return 'Waiting';
        case TaskStatus.STARTING:
          return 'Starting...';
        case TaskStatus.RUNNING:
          return 'Running';
        case TaskStatus.SUCCESS:
          return 'Success';
        case TaskStatus.ERROR:
          return 'Failed';
        case TaskStatus.STOPPING:
          return 'Stopping...';
        case TaskStatus.STOPPED:
          return 'Stopped';
        case TaskStatus.CONFIRMED:
          return 'Confirmed';
        case TaskStatus.WAITING_CONFIRMATION:
          return 'Waiting confirmation';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },

    getStatusColor(status) {
      switch (status) {
        case TaskStatus.WAITING:
          return '';
        case TaskStatus.STARTING:
          return 'warning';
        case TaskStatus.RUNNING:
          return 'primary';
        case TaskStatus.SUCCESS:
          return 'success';
        case TaskStatus.ERROR:
          return 'error';
        case TaskStatus.STOPPING:
          return '';
        case TaskStatus.STOPPED:
          return '';
        case TaskStatus.CONFIRMED:
          return 'warning';
        case TaskStatus.WAITING_CONFIRMATION:
          return 'warning';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },
  },
};
</script>
