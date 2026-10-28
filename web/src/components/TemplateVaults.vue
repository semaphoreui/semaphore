<template>
  <div class="pb-6" style="margin-top: -10px;" data-testid="templateVaults">
    <v-dialog
      v-model="editDialog"
      :scrim="false"
      content-class="NestedDialog"
      width="300"
    >
      <v-card
        :color="$vuetify.theme.current.dark ? '#212121' : 'white'"
        data-testid="templateVaultDialog"
      >
        <v-card-title data-testid="templateVaultDialog-heading"></v-card-title>
        <v-card-text class="pb-0">
          <v-form
            ref="form"
            lazy-validation
            v-if="editedVault != null"
            data-testid="templateVaultDialog-form"
          >
            <v-alert
              :model-value="!!formError"
              color="error"
              data-testid="templateVaultDialog-error"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="$t('vaultName')"
              placeholder="default"
              v-model.trim="editedVault.name"
              :rules="[v => this.vaultNameRules(v)]"
              data-testid="templateVaultDialog-name"
            />

            <v-select
              v-model="editedVault.type"
              :label="$t('type')"
              :items="vaultTypes"
              item-value="id"
              item-title="name"
              required
              :rules="[v => !!v  || $t('type_required')]"
              @update:model-value="updateType"
              data-testid="templateVaultDialog-type"
            />

            <v-select
              v-model="editedVault.vault_key_id"
              v-if="editedVault.type === 'password'"
              :label="$t('vaultPassword2')"
              :items="vaultKeys"
              item-value="id"
              item-title="name"
              required
              :rules="[(v) => !!v || $t('vaultRequired')]"
              data-testid="templateVaultDialog-vaultKeyId"
              :data-state="keys == null ? 'loading' : 'ready'"
            ></v-select>

            <v-text-field
              v-model="editedVault.script"
              v-if="editedVault.type === 'script'"
              :label="$t('vaultScript')"
              :rules="[v => this.vaultScriptRules(v)]"
              data-testid="templateVaultDialog-script"
            />

          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="editDialog = false"
            data-testid="templateVaultDialog-cancel"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="saveVault()"
            data-testid="templateVaultDialog-save"
          >
            {{ editedVaultIndex == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <fieldset style="padding: 0 10px 2px 10px;
                     border-width: 1px;
                     border-color: rgba(133, 133, 133, 0.4);
                     background-color: rgba(133, 133, 133, 0.1);
                     border-radius: 8px;
                     font-size: 12px;"
    >
      <legend style="padding: 0 3px;" data-testid="templateVaults-title">{{ $t('vaults') }}</legend>
      <v-chip-group column style="margin-top: -4px;">
        <v-chip
          v-for="(v, i) in modifiedVaults"
          closable
          @click:close="deleteVault(i)"
          :key="v.name"
          @click="editVault(i)"
          color="gray"
          :data-testid="`templateVaults-chip-${v.name || 'default'}`"
        >
          {{ v.name || 'default' }}
        </v-chip>
        <v-chip @click="editVault(null)" data-testid="templateVaults-add">
          + <span class="ml-1" v-if="modifiedVaults.length === 0">{{ $t('vaultAdd') }}</span>
        </v-chip>
      </v-chip-group>
    </fieldset>
  </div>
</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';

export default {
  props: {
    projectId: Number,
    vaults: Array,
  },
  emits: ['change'],
  watch: {
    vaults(val) {
      this.var = val || [];
    },
  },

  async created() {
    this.modifiedVaults = (this.vaults || []).map((v) => ({ ...v }));
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  data() {
    return {
      editDialog: null,
      editedVault: null,
      editedVaultIndex: null,
      modifiedVaults: null,
      formError: null,
      keys: null,
      vaultTypes: [{
        id: 'password',
        name: `${this.$t('vaultTypePassword')}`,
      }, {
        id: 'script',
        name: `${this.$t('vaultTypeScript')}`,
      }],
    };
  },

  computed: {
    vaultKeys() {
      if (this.keys == null) {
        return null;
      }
      return this.keys.filter((key) => ['login_password'].includes(key.type));
    },
  },

  methods: {
    vaultNameRules(v) {
      // Increase the max if we are editing an existing vault
      const max = this.editedVaultIndex == null ? 0 : 1;
      if (v == null || v === '') {
        if (this.modifiedVaults.filter((vault) => vault.name === null).length > max) {
          return this.$t('vaultNameDefault');
        }
      } else if (
        this.modifiedVaults.filter((vault) => vault.name === v.toLowerCase().trim()).length > max
      ) {
        return this.$t('vaultNameUnique');
      }
      return true;
    },

    vaultScriptRules(v) {
      if (v == null || v === '') {
        return this.$t('vaultScriptRequired');
      }
      if (v.match(/-client(\..*)?$/i) === null) {
        return this.$t('vaultScriptClientRequired');
      }
      return true;
    },

    updateType() {
      if (this.editedVault.type === 'password') {
        this.editedVault.script = null;
      } else {
        this.editedVault.vault_key_id = null;
      }
    },

    editVault(index) {
      this.editedVault = index != null ? { ...this.modifiedVaults[index] } : {
        name: null,
        vault_key_id: null,
        type: 'password',
        script: null,
      };
      this.editedVaultIndex = index;

      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }

      this.editDialog = true;
    },

    async saveVault() {
      this.formError = null;

      const { valid } = await this.$refs.form.validate();
      if (!valid) {
        return;
      }

      if (this.editedVault.name == null || this.editedVault.name === '') {
        this.editedVault.name = null;
      } else {
        this.editedVault.name = this.editedVault.name.toLowerCase().trim();
      }

      if (this.editedVaultIndex != null) {
        this.modifiedVaults[this.editedVaultIndex] = this.editedVault;
      } else {
        this.modifiedVaults.push(this.editedVault);
      }

      this.editDialog = false;
      this.editedVault = null;
      this.$emit('change', this.modifiedVaults);
    },

    deleteVault(index) {
      this.modifiedVaults.splice(index, 1);
      this.$emit('change', this.modifiedVaults);
    },
  },
};
</script>
