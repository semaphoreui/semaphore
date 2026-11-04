<template>
  <div class="permissions-list">
    <v-chip
      v-for="p in ROLE_PERMISSIONS[scope].filter(x => permissions & x.permission)"
      :key="p.permission"
      size="small"
      variant="flat"
      :color="p.color"
      :class="p.textColor && `text-${p.textColor}`"
      :data-testid="`permissionChips-${p.permission}`"
    >{{ $t(p.label) }}</v-chip>

    <span v-if="permissions === 0" class="text-medium-emphasis" data-testid="permissionChips-none">
      &mdash;
    </span>
  </div>
</template>

<script>
import { ROLE_PERMISSIONS } from '@/lib/constants';

export default {
  props: {
    permissions: {
      type: Number,
      required: true,
    },
    scope: {
      type: String,
      default: 'default',
    },
  },
  data() {
    return {
      ROLE_PERMISSIONS,
    };
  },
};
</script>

<style scoped>
.permissions-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

/* A small chip of Vuetify 2 is 24 px high (`$chip-sizes`), Vuetify 4 makes it 26 px. */
.v-chip.v-chip--size-small {
  --v-chip-height: 24px;
}
</style>
