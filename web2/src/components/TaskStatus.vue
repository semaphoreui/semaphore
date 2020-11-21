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
  RUNNING: 'running',
  SUCCESS: 'success',
  ERROR: 'error',
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
        case TaskStatus.RUNNING:
          return '';
        case TaskStatus.SUCCESS:
          return 'mdi-check-circle';
        case TaskStatus.ERROR:
          return 'mdi-information';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },

    humanizeStatus(status) {
      switch (status) {
        case TaskStatus.WAITING:
          return 'Waiting';
        case TaskStatus.RUNNING:
          return 'Running';
        case TaskStatus.SUCCESS:
          return 'Success';
        case TaskStatus.ERROR:
          return 'Failed';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },

    getStatusColor(status) {
      switch (status) {
        case TaskStatus.WAITING:
          return '';
        case TaskStatus.RUNNING:
          return 'primary';
        case TaskStatus.SUCCESS:
          return 'success';
        case TaskStatus.ERROR:
          return 'error';
        default:
          throw new Error(`Unknown task status ${status}`);
      }
    },
  },
};
</script>
