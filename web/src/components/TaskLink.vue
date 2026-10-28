<template>
  <span class="task-link">
    <v-icon
        v-if="status != null"
        size="small"
        class="mr-1"
        :color="statusColor"
        :icon="`mdi-${statusIcon}`"
        :data-testid="`taskLink-status-${status}`"
    />

    <span v-if="disabled" data-testid="taskLink-label">{{ label }}</span>

    <v-tooltip
        v-else
        location="right"
        max-width="350"
        transition="fade-transition"
        :disabled="!tooltip"
    >
      <template v-slot:activator="{ props }">
        <a
            v-bind="props"
            @click="showTaskLog()"
            :class="{'task-link-with-tooltip': tooltip}"
            data-testid="taskLink-open"
        >
          {{ label }}
        </a>
      </template>

      <span data-testid="taskLink-tooltip">{{ tooltip }}</span>

    </v-tooltip>
  </span>
</template>
<style lang="scss">
@use 'sass:map';
@use '@/scss/variables' as settings;

// Vuetify 2 coloured every link and gave it a pointer (`.v-application a`); Vuetify 3 does not,
// and this link has no href.
.task-link a {
  color: rgb(var(--v-theme-primary));
  cursor: pointer;
}

.task-link-with-tooltip {
  text-decoration: underline !important;
  text-decoration-style: dashed !important;
  text-decoration-color: gray !important;
}

a.task-link-with-tooltip {
  &:hover {
    text-decoration-style: solid !important;
    text-decoration-color: map.get(settings.$blue, 'darken-2') !important;
  }
}

</style>
<script>
import EventBus from '@/event-bus';

export default {
  props: {
    label: String,
    tooltip: String,
    taskId: Number,
    disabled: Boolean,
    status: String,
  },
  computed: {
    statusColor() {
      switch (this.status) {
        case 'success':
          return 'success';
        case 'error':
          return 'red';
        default:
          return 'gray';
      }
    },
    statusIcon() {
      switch (this.status) {
        case 'success':
          return 'check';
        case 'error':
          return 'close';
        default:
          return 'clock-time-three-outline';
      }
    },
  },
  methods: {
    showTaskLog() {
      EventBus.$emit('i-show-task', {
        taskId: this.taskId,
      });
    },
  },
};
</script>
