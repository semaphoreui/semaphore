<template>
  <EditDialog
    v-model="dialog"
    :max-width="1000"
    :hide-buttons="true"
    :expandable="true"
    no-body-paddings
    @close="onClose()"
    test-id="taskLogDialog"
  >
    <template v-slot:title={}>
      <div class="text-truncate" style="max-width: calc(100% - 36px);">
        <v-skeleton-loader
          v-if="template == null"
          type="button"
          style="display: inline-block; margin-right: 10px;"
        ></v-skeleton-loader>
        <router-link
          v-else
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/templates/${template ? template.id : null}`"
          @click="close()"
        >{{ template ? template.name : null }}</router-link>
        <v-icon>mdi-chevron-right</v-icon>
        <span class="breadcrumbs__item">{{ $t('task', {expr: item ? item.id : null}) }}</span>
      </div>
    </template>
    <template v-slot:form="{}">
      <TaskLogView v-if="item != null" :project-id="projectId" :item="item" />
      <v-skeleton-loader
        class="task-log-view__placeholder"
        v-else
        type="
            table-heading,
            image,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line,
            list-item-two-line"
      ></v-skeleton-loader>
    </template>
  </EditDialog>
</template>
<style lang="scss">
.task-log-view__placeholder {
  margin-left: 24px;
  margin-right: 24px;
  height: calc(100vh - 208px);
}
</style>
<script>
import TaskLogView from '@/components/TaskLogView.vue';
import EditDialog from '@/components/EditDialog.vue';
import ProjectMixin from '@/components/ProjectMixin';

export default {
  components: { EditDialog, TaskLogView },

  mixins: [ProjectMixin],

  props: {
    value: Boolean,
    projectId: Number,
    itemId: Number,
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.item = null;
      this.template = null;
      this.dialog = val;
      await this.loadData();
    },

    async itemId() {
      await this.loadData();
    },
  },

  data() {
    return {
      item: null,
      dialog: null,
      template: null,
    };
  },

  methods: {
    close() {
      this.dialog = false;
      this.item = null;
      this.template = null;
      this.onClose();
    },

    async loadData() {
      this.item = await this.loadProjectResource('tasks', this.itemId);
      this.template = await this.loadProjectResource('templates', this.item.template_id);
    },

    onClose() {
      this.$emit('close');
    },
  },
};
</script>
