<template>
  <NoProjects v-if="user && !user.can_create_project" :user="user"/>
  <div v-else data-testid="restoreProject-page">
    <v-toolbar flat >
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="restoreProject-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        data-testid="restoreProject-title"
      >{{ $t('restoreProject') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-divider />

    <div
      style="margin: auto; max-width: 600px; padding: 0 16px;"
      class="CenterToScreen"
    >
        <div class="project-settings-form">
        <div style="height: 300px;">
            <RestoreProjectForm item-id="new" ref="editForm" @save="onSave" />
        </div>

        <div class="text-right">
            <v-btn
              color="primary"
              @click="restoreProject()"
              data-testid="restoreProject-restore"
            >{{ $t('restore') }}</v-btn>
        </div>
        </div>
    </div>

  </div>
</template>
<style lang="scss">

</style>
<script>
import EventBus from '@/event-bus';
import RestoreProjectForm from '@/components/RestoreProjectForm.vue';
import NoProjects from '@/components/NoProjects.vue';

export default {
  components: { RestoreProjectForm, NoProjects },
  props: {
    user: Object,
  },
  data() {
    return {
    };
  },

  methods: {
    onSave(e) {
      EventBus.$emit('i-project', {
        action: 'restore',
        item: e.item,
      });
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async restoreProject() {
      await this.$refs.editForm.save();
    },
  },
};
</script>
