<template>
  <div class="d-flex justify-space-between align-center mb-2" data-testid="secretSource-toggle">
    <b style="font-size: 13px; margin-left: 5px" data-testid="secretSource-label">{{ label }}</b>
    <v-btn-toggle
        data-testid="secretSource-options"
        :value="value"
        @change="$emit('input', $event)"
        tile
        group
        mandatory
    >
      <v-btn
        v-for="source in sources"
        :key="source.value"
        :value="source.value"
        small
        class="ma-0"
        style="border-radius: 4px"
        :data-testid="`secretSource-${source.value}`"
      >
        {{ source.text }}
      </v-btn>
    </v-btn-toggle>
  </div>
</template>
<script>
export default {
  props: {
    /** Where the secret is taken from: 'database', 'env' or 'file'. */
    value: String,
    /** Name of the secret, shown on the left of the toggle. */
    label: String,
    disabled: Boolean,
  },

  emits: ['input'],

  data() {
    return {
      sources: [
        { value: 'database', text: 'Store in DB' },
        { value: 'env', text: 'From ENV' },
        { value: 'file', text: 'From File' },
      ],
    };
  },
};
</script>
