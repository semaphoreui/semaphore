<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>New Project</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <div class="project-settings-form">
      <div style="height: 220px;">
        <ProjectForm project-id="new" ref="editForm"/>
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="createProject()">Create</v-btn>
      </div>
    </div>

  </div>
</template>
<style lang="scss">

</style>
<script>
import EventBus from '@/event-bus';
import ProjectForm from '@/components/ProjectForm.vue';

export default {
  components: { ProjectForm },
  data() {
    return {
    };
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async createProject() {
      const item = await this.$refs.editForm.save();
      if (!item) {
        return;
      }
      EventBus.$emit('i-project', {
        action: 'new',
        item,
      });
    },
  },
};
</script>
