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
          <!-- Vue 2 kept the spaces around the label (whitespace `condense`); after the status icon
               the dashed underline starts with the space, Vue 3 drops it. -->
          {{ ' ' }}{{ label }}{{ ' ' }}
        </a>
      </template>

      <span data-testid="taskLink-tooltip">{{ tooltip }}</span>

    </v-tooltip>
  </span>
</template>
<style lang="scss">
@use 'sass:map';
@use '@/scss/variables' as settings;

.task-link-with-tooltip {
  text-decoration: underline !important;
  text-decoration-style: dashed !important;
  text-decoration-color: gray !important;
}

// An icon without a colour is in the icon colour of the theme, as in Vuetify 2
// (`.theme--light.v-icon`: .54 / white); Vuetify 4 gives it the text colour.
.task-link > .v-icon.v-theme--light:not([class*="text-"]) {
  color: rgba(0, 0, 0, .54);
}

.task-link > .v-icon.v-theme--dark:not([class*="text-"]) {
  color: #fff;
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
          // No colour of its own: the grey icon of the theme (style below), as `gray` gave in
          // Vuetify 2, which has no such colour.
          return undefined;
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
