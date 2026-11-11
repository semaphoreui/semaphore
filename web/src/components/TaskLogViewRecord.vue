<template>
  <div class="task-log-records__record" data-testid="taskLog-record">
    <div class="task-log-records__time" data-testid="taskLog-recordTime">
      {{ $filters.formatTime(source.time) }}
    </div>
    <!-- A row made of an AI agent's events (lib/agentLog.js). -->
    <TaskLogAgentRecord
      v-if="source.agent"
      :agent="source.agent"
      :running="running"
      @toggle="$emit('toggle')"
    />
    <div
      v-else
      class="task-log-records__output"
      data-testid="taskLog-recordOutput"
      v-html="$filters.formatLog(source.output)">
    </div>
  </div>
</template>
<script>
import TaskLogAgentRecord from '@/components/TaskLogAgentRecord.vue';

export default {
  components: { TaskLogAgentRecord },

  props: {
    index: { // index of current item
      type: Number,
    },
    source: { // here is: {uid: 'unique_1', text: 'abc'}
      type: Object,
      default() {
        return {};
      },
    },
    // Whether the task is still running.
    running: {
      type: Boolean,
      default: true,
    },
  },

  emits: ['toggle'],
};
</script>
