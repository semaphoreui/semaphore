<template>
  <v-skeleton-loader
    v-if="!isLoaded"
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
    lazy-validation
    v-model="formValid"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-if="premiumFeatures.project_runners"
      v-model="item.runner_tag"
      :items="runnerTags"
      :label="$t('runner_tag')"
      item-value="tag"
      item-text="tag"
      outlined
      dense
      :disabled="formSaving"
      :placeholder="$t('runner_tag')"
    ></v-autocomplete>

    <v-autocomplete
      v-model="item.ssh_key_id"
      :label="$t('userCredentials')"
      :items="keys"
      item-value="id"
      item-text="name"
      :rules="[v => !!v || $t('user_credentials_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-autocomplete
      v-model="item.become_key_id"
      :label="$t('sudoCredentialsOptional')"
      clearable
      :items="loginPasswordKeys"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
      outlined
      dense
    ></v-autocomplete>

    <v-select
      v-model="item.type"
      :label="$t('type')"
      :rules="[v => !!v || $t('type_required')]"
      :items="inventoryTypes"
      item-value="id"
      item-text="name"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-select>

    <v-text-field
      v-model.trim="item.inventory"
      :label="$t('pathToInventoryFile')"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
      v-if="item.type === 'file'"
      outlined
      dense
    ></v-text-field>

    <v-select
      v-model="item.repository_id"
      :label="$t('repository') + ' (Optional)'"
      clearable
      :items="repositories"
      item-value="id"
      item-text="name"
      :disabled="formSaving"
      v-if="item.type === 'file'"
      outlined
      dense
    ></v-select>

    <codemirror
      :class="{
        'InventoryEditor': true,
        'InventoryEditor--static': item.type === 'static',
        'InventoryEditor--static-yaml': item.type === 'static-yaml',
      }"
      :style="{ border: '1px solid lightgray' }"
      v-model.trim="item.inventory"
      :options="cmOptions"
      v-if="item.type === 'static' || item.type === 'static-yaml'"
      :placeholder="$t('enterInventory')"
    />

  </v-form>
</template>
<style>
.InventoryEditor .CodeMirror {
  height: 160px !important;
}

.v-dialog--fullscreen .InventoryEditor--static .CodeMirror {
  height: calc(100vh - 540px) !important;
}

.v-dialog--fullscreen .InventoryEditor--static-yaml .CodeMirror {
  height: calc(100vh - 600px) !important;
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';

import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/display/placeholder.js';

export default {
  mixins: [ItemFormBase],

  components: {
    codemirror,
  },

  props: {
    premiumFeatures: Object,
  },

  data() {
    return {
      cmOptions: {
        tabSize: 2,
        indentUnit: 2,
        mode: 'text/x-ini',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
        extraKeys: {
          Tab(cm) {
            // If something is selected, indent that selection
            if (cm.somethingSelected()) {
              cm.indentSelection('add');
            } else {
              // Otherwise, insert two spaces at the cursor
              cm.replaceSelection('  ', 'end');
            }
          },
        },
      },
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
    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
  },
};
</script>
