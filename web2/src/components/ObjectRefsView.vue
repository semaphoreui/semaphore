<template>
  <div>
    <div>{{ title }}</div>

    <div v-if="objectRefs.templates.length > 0">
      <div>Templates</div>
      <router-link
        v-for="t in objectRefs.templates"
        :key="t.id"
        :to="`/project/${projectId}/templates/${t.id}`">{{ t.name }}
      </router-link>
    </div>

    <div v-if="objectRefs.inventories.length > 0">
      <div>Inventories</div>
      <a v-for="referrer in objectRefs.inventories" :key="referrer.id">{{ referrer.name }}</a>
    </div>

    <div v-if="objectRefs.repositories.length > 0">
      <div>Repositories</div>
      <a v-for="referrer in objectRefs.repositories" :key="referrer.id">{{ referrer.name }}</a>
    </div>
  </div>
</template>
<script>
export default {
  props: {
    objectRefs: Object,
    projectId: Number,
    title: String,
  },
};
</script>
