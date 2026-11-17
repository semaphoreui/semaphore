<template>
  <div data-testid="inventoryVariables-page" :data-state="state">
    <YesNoDialog
      v-model="leaveDialog"
      test-id="inventoryVariablesLeaveDialog"
      :title="$t('inventory_tab_variables')"
      :text="$t('inventory_variables_leave')"
      @yes="confirmLeave()"
    />

    <v-progress-linear
      v-if="state === 'loading'"
      indeterminate
      color="primary-darken-2"
      bg-opacity=".3"
      data-testid="inventoryVariables-loading"
    ></v-progress-linear>

    <div v-else-if="inventory != null" class="pa-4">
      <v-alert
        :model-value="!!error"
        color="error"
        class="mb-4"
        data-testid="inventoryVariables-error"
      >{{ error }}</v-alert>

      <CodeMirror
        class="InventoryVariables__editor"
        :style="{ border: '1px solid lightgray' }"
        v-model="inventory.variables"
        minimal
        :tab-size="2"
        :extensions="cmExtensions"
        :disabled="!canManage || state === 'saving'"
        :readonly="!canManage"
        data-testid="inventoryVariables-editor"
      />

      <div v-if="canManage" class="d-flex justify-end mt-4">
        <v-btn
          color="primary"
          variant="elevated"
          :disabled="state === 'saving'"
          data-testid="inventoryVariables-save"
          @click="save()"
        >{{ $t('save') }}</v-btn>
      </div>
    </div>
  </div>
</template>
<style>
.InventoryVariables__editor .cm-editor {
  height: 360px !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import axios from 'axios';
import { markRaw } from 'vue';
import CodeMirror from 'vue-codemirror6';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import codeEditorExtensions, { CODEMIRROR5_LOOK } from '@/lib/codeEditor';
import PermissionsCheck from '@/components/PermissionsCheck';
import ProjectMixin from '@/components/ProjectMixin';
import YesNoDialog from '@/components/YesNoDialog.vue';

export default {
  mixins: [PermissionsCheck, ProjectMixin],

  components: { CodeMirror, YesNoDialog },

  props: {
    inventoryId: Number,
  },

  data() {
    return {
      inventory: null,
      error: null,
      state: 'loading',
      savedVariables: null,
      leaveDialog: false,
      pendingLeave: null,
      cmExtensions: markRaw([
        ...codeEditorExtensions({ language: 'yaml', tab: 'spaces' }),
        CODEMIRROR5_LOOK,
      ]),
    };
  },

  computed: {
    canManage() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    dirty() {
      return this.inventory != null && (this.inventory.variables || '') !== this.savedVariables;
    },
  },

  watch: {
    leaveDialog(open) {
      if (!open && this.pendingLeave) {
        this.pendingLeave(false);
        this.pendingLeave = null;
      }
    },
  },

  beforeRouteLeave(to, from, next) {
    if (!this.dirty) {
      next();
      return;
    }
    this.pendingLeave = next;
    this.leaveDialog = true;
  },

  mounted() {
    window.addEventListener('beforeunload', this.onBeforeUnload);
  },

  beforeUnmount() {
    window.removeEventListener('beforeunload', this.onBeforeUnload);
  },

  async created() {
    try {
      this.inventory = await this.loadProjectResource('inventory', this.inventoryId);
      this.savedVariables = this.inventory.variables || '';
      this.state = 'ready';
    } catch (err) {
      this.state = 'error';
      EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
    }
  },

  methods: {
    onBeforeUnload(ev) {
      if (this.dirty) {
        ev.preventDefault();
      }
    },

    confirmLeave() {
      const next = this.pendingLeave;
      this.pendingLeave = null;
      next();
    },

    async save() {
      this.error = null;
      this.state = 'saving';
      try {
        await axios.put(
          `/api/project/${this.projectId}/inventory/${this.inventoryId}/variables`,
          { variables: this.inventory.variables || '' },
        );
        this.savedVariables = this.inventory.variables || '';
        EventBus.$emit('i-snackbar', { color: 'success', text: this.$t('inventory_variables_saved') });
      } catch (err) {
        this.error = getErrorMessage(err);
      }
      this.state = 'ready';
    },
  },
};
</script>
