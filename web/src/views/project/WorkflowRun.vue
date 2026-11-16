<template>
  <WorkflowRunView
    :key="`${workflowId}-${runId}`"
    :project-id="projectId"
    :workflow-id="workflowId"
    :run-id="runId"
    :user-permissions="userPermissions"
    :is-admin="isAdmin"
  />
</template>

<script>
import WorkflowRunView from '@/components/WorkflowRunView.vue';

// The page of one workflow run, at the address tasks get as SEMAPHORE_WORKFLOW_URL. The links of
// the interface open the same view in the run panel instead (`?w=<workflow>/<run>`,
// WorkflowRunPanel.vue).
export default {
  components: { WorkflowRunView },
  // The router view hands every page the same set of props; the view takes the ones below.
  inheritAttrs: false,
  props: {
    projectId: Number,
    userPermissions: Number,
    isAdmin: Boolean,
  },
  computed: {
    workflowId() {
      return parseInt(this.$route.params.workflowId, 10);
    },
    runId() {
      return parseInt(this.$route.params.runId, 10);
    },
  },
};
</script>
