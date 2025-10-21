<template>
  <div class="permissions-list">
    <v-chip
      v-for="p in ROLE_PERMISSIONS[scope].filter(x => permissions & x.permission)"
      :key="p.permission"
      small
      :color="p.color"
      :text-color="p.textColor"
    >{{ $t(p.label) }}</v-chip>

    <span v-if="permissions === 0" class="text--secondary">
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
</style>
