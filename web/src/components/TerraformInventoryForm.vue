<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null && keys != null"
    class="TerraformInventoryForm"
    data-testid="tfWorkspace-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="tfWorkspace-error"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.inventory"
      label="Workspace name"
      :rules="[v => !!v || $t('path_required')]"
      required
      :disabled="formSaving"
      data-testid="tfWorkspace-inventory"
    ></v-text-field>

    <v-select
      v-model="item.ssh_key_id"
      label="SSH key for private modules *"
      :rules="[v => !!v || 'Key is required']"
      density="compact"
      required
      :items="keys"
      item-value="id"
      item-title="name"
      :disabled="formSaving"
      :menu-props="keyMenu"
      class="TerraformInventoryForm__key"
      data-testid="tfWorkspace-sshKeyId"
    ></v-select>

  </v-form>
</template>
<style lang="scss">
// An underlined field of Vuetify 2 is a 4 px top margin, 12 px for the label and the 32 px line;
// the 48 px field of Vuetify 4 holds the margin. Under the 16 px margin of the error alert that
// margin collapsed, so the fields stood 4 px higher (as RestoreProjectForm.vue).
.TerraformInventoryForm > .v-alert + .v-input {
  margin-top: -4px;
}

// The dense underlined select of Vuetify 2 (`v-select dense`): 4 px over the field
// (`.v-text-field` margin-top), a 26 px line (Vuetify 4 compact: 32 px under no margin) with the
// selection 2 px down, the arrow at its top, the resting label centred on the line and the floating
// one 18 px over the resting one, scaled to .75.
.TerraformInventoryForm__key {
  margin-top: 4px;

  .v-field__input {
    min-height: 26px;
    padding-top: 2px;
  }

  .v-field__append-inner {
    padding-top: 0;
  }

  .v-field-label {
    top: 2px;
  }

  .v-field-label.v-field-label--floating {
    top: 1px;
  }
}
</style>
<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */

import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  props: {
    namePrefix: String,
    type: String,
    templateId: Number,
  },

  components: {
  },

  data() {
    return {
      keys: null,
      // The list over the field from its top edge (the default of plugins/vuetify.js), at most
      // 304 px high (Vuetify 2 `menuProps`; Vuetify 4: 310).
      keyMenu: { location: 'top', origin: 'top', maxHeight: 304 },
    };
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    async beforeSave() {
      this.item.name = this.namePrefix + this.item.inventory;
      this.item.type = this.type;
      this.item.template_id = this.templateId;
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
