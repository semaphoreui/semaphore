<template>
  <v-text-field
    v-model="days"
    class="ScheduleOffsetField"
    type="number"
    :label="label"
    :aria-label="label ? undefined : $t('scheduleOffset')"
    :min="-MAX_OFFSET_DAYS"
    :max="MAX_OFFSET_DAYS"
    :suffix="$tc('scheduleOffsetDays', Math.abs(days))"
    :hint="hint"
    persistent-hint
    :disabled="disabled"
    outlined
    dense
    @change="commit"
  >
    <template v-slot:prepend-inner>
      <v-btn
        icon
        small
        :disabled="disabled"
        :aria-label="$t('scheduleOffsetEarlier')"
        @click="commit(Number(days) - 1)"
      >
        <v-icon>mdi-minus</v-icon>
      </v-btn>
    </template>

    <template v-slot:append>
      <v-btn
        icon
        small
        :disabled="disabled"
        :aria-label="$t('scheduleOffsetLater')"
        @click="commit(Number(days) + 1)"
      >
        <v-icon>mdi-plus</v-icon>
      </v-btn>
    </template>
  </v-text-field>
</template>

<style lang="scss">
.ScheduleOffsetField {
  max-width: 240px;

  input {
    text-align: center;
    -moz-appearance: textfield;
  }

  input::-webkit-outer-spin-button,
  input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
}
</style>

<script>
import { clampOffsetDays, MAX_OFFSET_DAYS } from '@/lib/cronPresets';

export default {
  props: {
    value: {
      type: Number,
      default: 0,
    },
    label: String,
    hint: String,
    disabled: Boolean,
  },

  data() {
    return {
      days: this.value,
      MAX_OFFSET_DAYS,
    };
  },

  watch: {
    value(val) {
      this.days = val;
    },
  },

  methods: {
    // Typed input is applied on change rather than on every keystroke, so a
    // lone "-" never reaches the server.
    commit(input) {
      const days = clampOffsetDays(input);
      this.days = days;

      if (days !== this.value) {
        this.$emit('input', days);
      }
    },
  },
};
</script>
