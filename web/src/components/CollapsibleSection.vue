<template>
  <div>
    <div>
      <v-btn
        variant="text"
        size="small"
        class="px-1 text-none"
        @click="$emit('update:modelValue', !modelValue)"
        data-testid="collapsibleSection-toggle"
      >
        <v-icon
          start
          size="small"
          class="mr-1 CollapsibleSection__chevron"
          :class="{ 'CollapsibleSection__chevron--open': modelValue }"
        >mdi-chevron-right</v-icon>
        {{ title }}
        <v-chip
          v-if="badge"
          size="x-small"
          class="ml-2"
          data-testid="collapsibleSection-badge"
        >{{ badge }}</v-chip>
      </v-btn>
    </div>

    <DropdownCard
      v-if="modelValue"
      class="mt-4"
      data-testid="collapsibleSection-body"
    >
      <slot />
    </DropdownCard>
  </div>
</template>

<style lang="scss">
.CollapsibleSection__chevron {
  transition: transform 0.2s;

  &--open {
    transform: rotate(90deg);
  }
}
</style>

<script>
import DropdownCard from '@/components/DropdownCard.vue';

export default {
  components: { DropdownCard },
  props: {
    modelValue: Boolean,
    title: String,
    badge: String,
  },
  emits: ['update:modelValue'],
};
</script>
