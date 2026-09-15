<template>
  <div>
    <div>
      <v-btn
        text
        small
        class="px-1 text-none"
        @click="$emit('input', !value)"
      >
        <v-icon
          left
          small
          class="mr-1 CollapsibleSection__chevron"
          :class="{ 'CollapsibleSection__chevron--open': value }"
        >mdi-chevron-right</v-icon>
        {{ title }}
        <v-chip v-if="badge" x-small class="ml-2">{{ badge }}</v-chip>
      </v-btn>
    </div>

    <DropdownCard v-if="value" class="mt-4">
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
    value: Boolean,
    title: String,
    badge: String,
  },
};
</script>
