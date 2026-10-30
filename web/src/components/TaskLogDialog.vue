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
    <template v-slot:title>
      <div
        class="text-truncate"
        style="max-width: calc(100% - 36px);"
        data-testid="taskLog-title"
        :data-state="item == null || template == null ? 'loading' : 'ready'"
      >
        <!-- The button bone fills the loader; an inline-block loader without a width shrinks it
             to nothing. The bone is 64 × 36 without margins, as in Vuetify 2 (style below). -->
        <v-skeleton-loader
          v-if="template == null"
          class="task-log-dialog__titleLoader"
          data-testid="taskLog-titleLoading"
          type="button"
          width="64"
          style="display: inline-block; margin-right: 10px; vertical-align: middle;"
        ></v-skeleton-loader>
        <router-link
          v-else
          class="breadcrumbs__item breadcrumbs__item--link"
          :to="`/project/${projectId}/templates/${template ? template.id : null}`"
          data-testid="taskLog-templateLink"
        >{{ template ? template.name : null }}</router-link>
        <v-icon class="task-log-dialog__chevron">mdi-chevron-right</v-icon>
        <span
          class="breadcrumbs__item"
          data-testid="taskLog-taskTitle"
        >{{ $t('task', {expr: item ? item.id : null}) }}</span>
      </div>
    </template>
    <template v-slot:form>

      <TaskLogView
        v-if="item != null"
        :project-id="projectId"
        :item="item"
        :system-info="systemInfo"
      />

      <v-skeleton-loader
        class="task-log-view__placeholder"
        v-else
        data-testid="taskLog-loading"
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
  height: calc(100dvh - 208px);
}

// The chevron in the icon colour of Vuetify 2 (`.theme--light.v-icon`: .54, white in the dark
// theme), not the colour of the title.
.task-log-dialog__chevron.v-theme--light {
  color: rgba(0, 0, 0, .54);
}

.task-log-dialog__chevron.v-theme--dark {
  color: #fff;
}

// The skeleton of the title: the button bone of Vuetify 2 (`$skeleton-loader-button-*`: 64 × 36,
// no margin); Vuetify 4 puts 16 px around it, the title row grew by 32 px.
.task-log-dialog__titleLoader .v-skeleton-loader__button {
  max-width: none;
  margin: 0;
}

// The skeleton of the body as Vuetify 2 drew it (`VSkeletonLoader.sass`): bones in blocks, a table
// heading of 16 px padding with a 24 px heading, a 200 px image, two-line items 72 px high; texts
// 12 px with 6 px under each.
.task-log-view__placeholder.v-skeleton-loader {
  display: block;

  .v-skeleton-loader__text {
    flex: 1 0 auto;
    height: 12px;
    margin: 0 0 6px;
    max-width: none;
  }

  .v-skeleton-loader__table-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;

    > .v-skeleton-loader__heading {
      height: 24px;
      margin: 0;
      width: 45%;
      max-width: 15%;
    }

    > .v-skeleton-loader__text {
      max-width: 40%;
    }
  }

  .v-skeleton-loader__image {
    height: 200px;
  }

  .v-skeleton-loader__list-item-two-line {
    display: flex;
    flex-wrap: wrap;
    align-content: center;
    align-items: center;
    height: 72px;
    margin: 0;
    padding: 0 16px;
  }

  .v-skeleton-loader__sentences {
    display: block;
    flex: 1 0 auto;

    > .v-skeleton-loader__text:nth-child(2) {
      max-width: 70%;
      margin-bottom: 0;
    }
  }
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
    modelValue: Boolean,
    projectId: Number,
    itemId: Number,
    systemInfo: Object,
  },

  emits: ['update:modelValue', 'close'],

  watch: {
    async dialog(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
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
      if (this.itemId == null) {
        return;
      }
      this.item = await this.loadProjectResource('tasks', this.itemId);
      this.template = await this.loadProjectResource('templates', this.item.template_id);
    },

    onClose() {
      this.$emit('close');
    },
  },
};
</script>
