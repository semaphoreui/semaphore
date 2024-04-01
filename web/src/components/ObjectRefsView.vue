<template>
  <div class="object-refs-view">
    <v-alert
      type="warning"
    >
      {{ $t('theCantBeDeletedBecauseItUsedByTheResourcesBelow', {objectTitle: objectTitle}) }}
    </v-alert>
    <div
      v-for="s in sections"
      class="object-refs-view__section"
      :key="s.slug"
    >
      <div class="object-refs-view__section-title">
        <v-icon small class="mr-2">mdi-{{ s.icon }}</v-icon>{{ s.title }}:
      </div>

      <div class="ml-6">
        <span v-for="t in objectRefs[s.slug]" class="object-refs-view__link-wrap" :key="t.id">
          <router-link
            :to="`/project/${projectId}/templates/${t.id}`"
            class="object-refs-view__link">{{ t.name }}</router-link>
        </span>
      </div>
    </div>
  </div>
</template>
<style lang="scss">
.object-refs-view__section {
  margin-bottom: 10px;
}

.object-refs-view__link-wrap + .object-refs-view__link-wrap {
  &:before {
    content: ", ";
  }
}
</style>
<script>
export default {
  props: {
    objectRefs: Object,
    projectId: Number,
    objectTitle: String,
  },
  computed: {
    sections() {
      return [{
        slug: 'templates',
        title: 'Templates',
        icon: 'check-all',
      }, {
        slug: 'inventories',
        title: 'Inventories',
        icon: 'monitor-multiple',
      }, {
        slug: 'repositories',
        title: 'Repositories',
        icon: 'git',
      }].filter((s) => this.objectRefs[s.slug].length > 0);
    },
  },
};
</script>
