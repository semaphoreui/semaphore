<template>
  <div>
    <div>
      <v-btn
        variant="text"
        size="small"
        class="px-1 text-none CollapsibleSection__toggle"
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
          class="ml-2 CollapsibleSection__badge"
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
// The look of Vuetify 2: the text of a button without a colour in the text colour of the theme
// (`.theme--light.v-btn`: .87), not in the .6 of the card text around it.
.CollapsibleSection__toggle.v-theme--light {
  color: rgba(0, 0, 0, .87);
}

.CollapsibleSection__toggle.v-theme--dark {
  color: #fff;
}

// The chevron: a small icon (16 px) in the 18 px box of an icon beside the button text,
// pulled 4 px to the edge (`.v-btn .v-icon--left`).
.CollapsibleSection__chevron.v-icon {
  font-size: 16px;
  width: 18px;
  height: 18px;
  margin-inline-start: -4px;
}

// The badge: an x-small chip of Vuetify 2 is 16 px high and takes the weight of the button text.
.CollapsibleSection__badge.v-chip {
  --v-chip-height: 16px;
  font-weight: inherit;
}

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
