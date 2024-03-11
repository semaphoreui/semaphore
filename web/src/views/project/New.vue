<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ projectType === 'premium' ? 'Buy Premium License' : $t('newProject') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <div v-if="projectType === 'premium'" class="project-settings-form">
      <div style="height: 150px;">
        <PremiumLicenseProjectForm item-id="new" ref="editForm" @save="onSave" />
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="createProject()">Continue</v-btn>
      </div>
    </div>

    <div v-else class="project-settings-form">
      <div style="height: 300px;">
        <ProjectForm item-id="new" ref="editForm" @save="onSave" />
      </div>

      <div class="text-right">
        <v-btn
          color="success" class="mr-3" @click="createDemoProject()"
        >{{ $t('CreateDemoProject') }}</v-btn>

        <v-btn color="primary" @click="createProject()">{{ $t('create') }}</v-btn>
      </div>
    </div>

  </div>
</template>
<style lang="scss">

</style>
<script>
import EventBus from '@/event-bus';
import ProjectForm from '@/components/ProjectForm.vue';
import PremiumLicenseProjectForm from '@/components/PremiumLicenseProjectForm.vue';

export default {
  components: { PremiumLicenseProjectForm, ProjectForm },

  data() {
    return {
    };
  },

  props: {
    projectType: String,
  },

  methods: {
    onSave(e) {
      EventBus.$emit('i-project', {
        action: 'new',
        item: e.item,
      });
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async createProject() {
      this.$refs.editForm.item.type = this.projectType;
      await this.$refs.editForm.save();
    },

    async createDemoProject() {
      await this.$refs.editForm.save({
        demo: true,
      });
    },
  },
};
</script>
