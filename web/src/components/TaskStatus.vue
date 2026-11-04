<template>
  <v-chip
    v-if="status"
    style="font-weight: bold;"
    class="task-status"
    :class="{'task-status--no-color': !getStatusColor(status)}"
    :color="getStatusColor(status)"
  >
    <v-icon
      v-if="status !== 'running'"
      start
      size="24"
      :icon="getStatusIcon(status)"
      :data-testid="`taskStatus-${status}`"
    />
    <IndeterminateProgressCircular
      v-else
      style="margin-left: -5px;"
      :data-testid="`taskStatus-${status}`"
    />
    {{ humanizeStatus(status) }}
  </v-chip>
</template>
<style lang="scss">
// Vuetify 2 look: the chip is filled (VChip default `flat`; without a colour — the grey chip of
// Vuetify 2), its icon 24 px (`$chip-icon-size`; Vuetify 4 scales it to 18/21) and, on a chip
// without a colour, in the icon colour of the theme (`.theme--light.v-icon`: .54 / white), not in
// the text colour. The text line is 20 px (`$chip-line-height`), not the 21 px around the chip.
.v-chip.task-status {
  line-height: 20px;
}

.v-chip.task-status--no-color.v-theme--light .v-icon {
  color: rgba(0, 0, 0, .54);
}

.v-chip.task-status--no-color.v-theme--dark .v-icon {
  color: #fff;
}
</style>
<script>
import IndeterminateProgressCircular from '@/components/IndeterminateProgressCircular.vue';

const TaskStatus = Object.freeze({
  WAITING: 'waiting',
  STARTING: 'starting',
  WAITING_CONFIRMATION: 'waiting_confirmation',
  CONFIRMED: 'confirmed',
  REJECTED: 'rejected',
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
