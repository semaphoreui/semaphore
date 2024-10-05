<template>
  <div class="pb-6">
    <v-dialog
      v-model="editDialog"
      hide-overlay
      width="300"
    >
      <v-card :color="$vuetify.theme.dark ? '#212121' : 'white'">
        <v-card-title></v-card-title>
        <v-card-text class="pb-0">
          <v-form
            ref="form"
            lazy-validation
            v-if="editedVault != null"
          >
            <v-alert
              :value="formError"
              color="error"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="$t('vaultName')"
              placeholder="default"
              v-model.trim="editedVault.name"
              :rules="[v => this.vaultNameRules(v)]"
            />

            <v-select
              v-model="editedVault.vault_key_id"
              :label="$t('vaultPassword2')"
              :items="vaultKeys"
              item-value="id"
              item-text="name"
              required
              :rules="[(v) => !!v || $t('vaultRequired')]"
            ></v-select>

          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue darken-1"
            text
            @click="editDialog = false"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="blue darken-1"
            text
            @click="saveVault()"
          >
            {{ editedVaultIndex == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <fieldset style="padding: 0 10px 2px 10px;
                     border: 1px solid rgba(0, 0, 0, 0.38);
                     border-radius: 4px;
                     font-size: 12px;"
              :style="{
                       'border-color': $vuetify.theme.dark ?
                         'rgba(200, 200, 200, 0.38)' :
                         'rgba(0, 0, 0, 0.38)'
                     }">
      <legend style="padding: 0 3px;">{{ $t('vaults') }}</legend>
      <v-chip-group column style="margin-top: -4px;">
        <v-chip
          v-for="(v, i) in modifiedVaults"
          close
          @click:close="deleteVault(i)"
          :key="v.name"
          @click="editVault(i)"
          color="gray"
        >
          {{ v.name || 'default' }}
        </v-chip>
        <v-chip @click="editVault(null)">
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
      if (v == null || v === '') {
        if (this.modifiedVaults.some((vault) => vault.name === null)) {
          return this.$t('vaultNameDefault');
        }
      } else if (this.modifiedVaults.some((vault) => vault.name === v.toLowerCase().trim())) {
        return this.$t('vaultNameUnique');
      }
      return true;
    },

    editVault(index) {
      this.editedVault = index != null ? { ...this.modifiedVaults[index] } : {
        name: null,
        vault_key_id: null,
      };
      this.editedVaultIndex = index;

      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }

      this.editDialog = true;
    },

    saveVault() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
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
