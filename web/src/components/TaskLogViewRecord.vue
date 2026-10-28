<template>
  <div class="task-log-records__record" data-testid="taskLog-record">
    <div class="task-log-records__time" data-testid="taskLog-recordTime">
      {{ source.time | formatTime }}
    </div>
    <div
      class="task-log-records__output"
      data-testid="taskLog-recordOutput"
      v-html="$options.filters.formatLog(source.output)">
    </div>
  </div>
</template>
<script>
export default {
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
  },
};
</script>
