<template>
  <div
    class="pa-8 text-center text-medium-emphasis"
    data-testid="workflowStats-page"
    data-state="empty"
  >
    {{ $t('workflowStatsComingSoon') }}
  </div>
</template>

<script>
export default {
  props: {
    projectId: Number,
    workflow: Object,
  },
};
</script>
