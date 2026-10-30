<template>
  <v-skeleton-loader
    v-if="!isLoaded"
    class="InventoryForm__loader"
    data-testid="inventoryForm-loading"
    :data-state="formError ? 'error' : 'loading'"
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
  <v-form
    v-else
    ref="form"
    v-model="formValid"
    data-testid="inventoryForm-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="inventoryForm-error"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-name"
    ></v-text-field>

    <v-autocomplete
      v-if="features.project_runners"
      v-model="item.runner_tag"
      :items="runnerTags"
      :label="$t('runner_tag')"
      item-value="tag"
      item-title="tag"
      variant="outlined"
      density="compact"
      clearable
      :disabled="formSaving"
      :placeholder="$t('runner_tag')"
      data-testid="inventoryForm-runnerTag"
    ></v-autocomplete>

    <v-autocomplete
      v-model="item.ssh_key_id"
      :label="$t('userCredentials')"
      :items="keys"
      item-value="id"
      item-title="name"
      :rules="[v => !!v || $t('user_credentials_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-sshKeyId"
    ></v-autocomplete>

    <v-autocomplete
      v-model="item.become_key_id"
      :label="$t('sudoCredentialsOptional')"
      clearable
      :items="loginPasswordKeys"
      item-value="id"
      item-title="name"
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-becomeKeyId"
    ></v-autocomplete>

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :rules="[v => !!v || $t('type_required')]"
      :items="inventoryTypes"
      item-value="id"
      item-title="name"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-type"
    ></v-select>

    <v-text-field
      v-model.trim="item.inventory"
      :label="$t('pathToInventoryFile')"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
      v-if="item.type === 'file'"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-inventoryPath"
    ></v-text-field>

    <v-select
      v-model="item.repository_id"
      :label="$t('repository') + ' (Optional)'"
      clearable
      :items="repositories"
      item-value="id"
      item-title="name"
      :disabled="formSaving"
      v-if="item.type === 'file'"
      variant="outlined"
      density="compact"
      data-testid="inventoryForm-repositoryId"
    ></v-select>

    <div
      style="position: relative"
      v-if="item.type === 'static' || item.type === 'static-yaml'"
    >
      <CodeMirror
        :class="{
        'InventoryEditor': true,
        'InventoryEditor--static': item.type === 'static',
        'InventoryEditor--static-yaml': item.type === 'static-yaml',
      }"
        :style="{ border: '1px solid lightgray' }"
        v-model.trim="item.inventory"
        minimal
        :tab-size="2"
        :extensions="cmExtensions"
        :placeholder="$t('enterInventory')"
        data-testid="inventoryForm-inventoryCode"
      />

      <RichEditor
        v-model.trim="item.inventory"
        type="ini"
        test-id="inventoryForm-inventoryEditor"
        style="
              position: absolute;
              right: 0;
              top: 0;
              margin: 10px;
            "
      />

    </div>

  </v-form>
</template>
<style>
.InventoryEditor .cm-editor {
  height: 160px !important;
}

.v-dialog--fullscreen .InventoryEditor--static .cm-editor {
  height: calc(100dvh - 540px) !important;
}

.v-dialog--fullscreen .InventoryEditor--static-yaml .cm-editor {
  height: calc(100dvh - 600px) !important;
}

/* The loading skeleton of Vuetify 2 (`VSkeletonLoader.sass`, as TemplateForm.vue): bones in
   blocks, a table heading of 16 px padding with a 24 px heading, a 200 px image, two-line items
   72 px high; texts 12 px with 6 px under each, the second line of an item 70 % wide. */
.InventoryForm__loader.v-skeleton-loader {
  display: block;
}

.InventoryForm__loader .v-skeleton-loader__text {
  flex: 1 0 auto;
  height: 12px;
  margin: 0 0 6px;
  max-width: none;
}

.InventoryForm__loader .v-skeleton-loader__table-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
}

.InventoryForm__loader .v-skeleton-loader__table-heading > .v-skeleton-loader__heading {
  height: 24px;
  margin: 0;
  width: 45%;
  max-width: 15%;
}

.InventoryForm__loader .v-skeleton-loader__table-heading > .v-skeleton-loader__text {
  max-width: 40%;
}

.InventoryForm__loader .v-skeleton-loader__image {
  height: 200px;
}

.InventoryForm__loader .v-skeleton-loader__list-item-two-line {
  display: flex;
  flex-wrap: wrap;
  align-content: center;
  align-items: center;
  height: 72px;
  margin: 0;
  padding: 0 16px;
}

.InventoryForm__loader .v-skeleton-loader__sentences {
  display: block;
  flex: 1 0 auto;
}

.InventoryForm__loader .v-skeleton-loader__sentences > .v-skeleton-loader__text:nth-child(2) {
  max-width: 70%;
  margin-bottom: 0;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { markRaw } from 'vue';
import CodeMirror from 'vue-codemirror6';
import { EditorView } from '@codemirror/view';
import codeEditorExtensions from '@/lib/codeEditor';
import RichEditor from '@/components/RichEditor.vue';

// The editor as CodeMirror 5 drew it in Vuetify 2 (`codemirror.css`): black text on white in the
// line height of the card text (22 px, not 1.4), lines padded 4 px, a 30 px gutter of #f7f7f7 with
// numbers in #999. A theme, not a style block: CodeMirror 6 mounts its own styles outside the CSS
// layers, where they win over any layered rule.
const CODEMIRROR5_LOOK = EditorView.theme({
  '&': {
    color: '#000',
    backgroundColor: '#fff',
  },
  '.cm-scroller': {
    lineHeight: 'inherit',
  },
  '.cm-line': {
    padding: '0 4px',
  },
  '.cm-gutters': {
    backgroundColor: '#f7f7f7',
    color: '#999',
  },
  '.cm-lineNumbers .cm-gutterElement': {
    minWidth: '29px',
  },
});

export default {
  mixins: [ItemFormBase],

  components: {
    RichEditor,
    CodeMirror,
  },

  props: {
    features: Object,
  },

  data() {
    return {
      cmExtensions: markRaw([
        ...codeEditorExtensions({ tab: 'spaces' }),
        CODEMIRROR5_LOOK,
      ]),
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'static-yaml',
        name: 'Static YAML',
      }, {
        id: 'file',
        name: 'File',
      }],
      keys: null,
      repositories: null,
      runnerTags: null,
    };
  },

  computed: {
    loginPasswordKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => key.type === 'login_password');
    },
    isLoaded() {
      return this.item != null && this.keys != null;
    },
  },

  async created() {
    [
      this.keys,
      this.repositories,
      this.runnerTags,
    ] = await Promise.all([
      this.loadProjectResources('keys'),
      this.loadProjectResources('repositories'),
      this.loadProjectResources('runner_tags'),
    ]);
  },

  methods: {
    beforeSave() {
      // The repository selector is only shown for "file" inventories; a value
      // left over from a previous type must not be submitted.
      if (this.item.type !== 'file') {
        this.item.repository_id = null;
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
