<template>
  <div class="task-log-records__record">
    <div class="task-log-records__time">
      {{ source.time | formatTime }}
    </div>
    <div class="task-log-records__output" v-html="$options.filters.formatLog(source.output)">
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
