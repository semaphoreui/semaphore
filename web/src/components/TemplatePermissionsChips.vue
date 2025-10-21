<template>
  <div class="permissions-list">
    <v-chip
      v-for="p in PERMISSIONS[scope].filter(x => permissions & x.permission)"
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
const PERMISSIONS = {
  default: [{
    permission: 1,
    label: 'canRunProjectTasks',
    color: 'blue',
    textColor: 'white',
  }, {
    permission: 2,
    label: 'canUpdateProject',
    color: 'green',
    textColor: 'white',
  }, {
    permission: 4,
    label: 'canManageProjectResources',
    color: 'orange',
    textColor: 'white',
  }, {
    permission: 8,
    label: 'canManageProjectUsers',
    color: 'red',
    textColor: 'white',
  }],
  template: [{
    permission: 1,
    label: 'Can run tasks',
    color: 'blue',
    textColor: 'white',
  }, {
    permission: 4,
    label: 'Can update the template',
    color: 'orange',
    textColor: 'white',
  }],
};
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
      PERMISSIONS,
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
