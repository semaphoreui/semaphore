<template>
  <span
    v-if="segments.length > 0"
    class="view-template-stats"
    data-testid="viewTemplateStats"
  >
    <v-chip
      v-for="segment in segments"
      :key="segment.key"
      size="x-small"
      class="view-template-stats__segment"
      :color="segment.color"
      :title="`${segment.title}: ${segment.count}`"
      :data-testid="`viewTemplateStats-${segment.key}`"
    >{{ segment.count }}</v-chip>
  </span>
</template>
<style lang="scss">
// One chip cut into coloured segments: only the outer corners are round.
.view-template-stats {
  display: inline-flex;
  margin-left: 8px;
  vertical-align: middle;
}

.view-template-stats .v-chip.view-template-stats__segment {
  border-radius: 0;
  font-weight: bold;
  letter-spacing: normal;
  // the chips are a part of the tab link they sit in
  cursor: inherit;
  padding: 0 7px;
}

.view-template-stats__segment {
  opacity: 0.5;
}

.v-tab--selected .view-template-stats__segment {
  opacity: 0.8;
}

.view-template-stats .v-chip.view-template-stats__segment:first-child {
  border-start-start-radius: 10px;
  border-end-start-radius: 10px;
}

.view-template-stats .v-chip.view-template-stats__segment:last-child {
  border-start-end-radius: 10px;
  border-end-end-radius: 10px;
}
</style>
<script>
// The counts of a view in the order they are shown, with the colour of the task status they
// stand for (TaskStatus.vue): no colour is the grey chip of a stopped task.
const SEGMENTS = [
  { key: 'success', color: 'success', title: (vm) => vm.$t('status_success') },
  { key: 'failed', color: 'error', title: (vm) => vm.$t('status_failed') },
  {
    key: 'stopped',
    color: undefined,
    title: (vm) => `${vm.$t('status_stopped')} / ${vm.$t('notLaunched')}`,
  },
  { key: 'confirming', color: 'warning', title: (vm) => vm.$t('status_confirming') },
  { key: 'running', color: 'primary', title: (vm) => vm.$t('status_running') },
];

/**
 * Template counts of a view by the status of the last task, as one chip of joined coloured
 * segments. A count of zero has no segment; without any count nothing is drawn.
 */
export default {
  props: {
    // `template_stats` of a view: { success, failed, stopped, confirming, running }
    stats: Object,
  },

  computed: {
    segments() {
      const stats = this.stats || {};
      return SEGMENTS
        .filter((segment) => stats[segment.key] > 0)
        .map((segment) => ({
          key: segment.key,
          color: segment.color,
          title: segment.title(this),
          count: stats[segment.key],
        }));
    },
  },
};
</script>
