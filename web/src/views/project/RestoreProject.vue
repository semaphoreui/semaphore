<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('restoreProject') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <div class="project-settings-form">
      <div style="height: 300px;">
        <RestoreProjectForm item-id="new" ref="editForm" @save="onSave" />
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="restoreProject()">{{ $t('restore') }}</v-btn>
      </div>
    </div>

  </div>
</template>
<style lang="scss">

</style>
<script>
import EventBus from '@/event-bus';
import RestoreProjectForm from '@/components/RestoreProjectForm.vue';

export default {
  components: { RestoreProjectForm },
  data() {
    return {
    };
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async restoreProject() {
      await this.$refs.editForm.save();
    },
  },
};
</script>
