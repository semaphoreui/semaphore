<template>
  <div class="pa-8 text-center text--secondary">
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
