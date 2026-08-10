<template>
  <div class="pb-6" style="margin-top: -10px;">
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
            v-if="editedCredential != null"
          >
            <v-alert
              :value="formError"
              color="error"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="$t('submoduleCredentialHost')"
              placeholder="gitserver.example.com"
              v-model.trim="editedCredential.host"
              :rules="[v => !!v || $t('submoduleCredentialHostRequired')]"
            />

            <v-select
              v-model="editedCredential.access_key_id"
              :label="$t('accessKey')"
              :items="keys"
              item-value="id"
              item-text="name"
              required
              :rules="[v => !!v || $t('key_required')]"
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
            :loading="formSaving"
            @click="saveCredential()"
          >
            {{ editedCredentialId == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <fieldset v-if="hasRepository" style="padding: 0 10px 2px 10px;
                     border: 1px solid rgba(0, 0, 0, 0.38);
                     border-radius: 4px;
                     font-size: 12px;"
              :style="{
                       'border-color': $vuetify.theme.dark ?
                         'rgba(200, 200, 200, 0.38)' :
                         'rgba(0, 0, 0, 0.38)'
                     }">
      <legend style="padding: 0 3px;">{{ $t('submoduleCredentials') }}</legend>
      <v-chip-group column style="margin-top: -4px;">
        <v-chip
          v-for="c in credentials"
          close
          @click:close="deleteCredential(c)"
          :key="c.id"
          @click="editCredential(c)"
          color="gray"
        >
          {{ c.host }}
        </v-chip>
        <v-chip @click="editCredential(null)">
          + <span
            class="ml-1"
            v-if="credentials.length === 0"
          >{{ $t('submoduleCredentialAdd') }}</span>
        </v-chip>
      </v-chip-group>
    </fieldset>
    <div v-else class="caption" style="opacity: 0.6;">
      {{ $t('saveRepositoryFirstForSubmoduleCredentials') }}
    </div>
  </div>
</template>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: [Number, String],
    repositoryId: [Number, String],
    keys: Array,
  },

  data() {
    return {
      credentials: [],
      editDialog: null,
      editedCredential: null,
      editedCredentialId: null,
      formError: null,
      formSaving: false,
    };
  },

  computed: {
    hasRepository() {
      return this.repositoryId != null && this.repositoryId !== 'new';
    },
  },

  watch: {
    repositoryId() {
      this.loadCredentials();
    },
  },

  async created() {
    await this.loadCredentials();
  },

  methods: {
    baseUrl() {
      return `/api/project/${this.projectId}/repositories/${this.repositoryId}/submodule_credentials`;
    },

    async loadCredentials() {
      if (!this.hasRepository) {
        this.credentials = [];
        return;
      }

      try {
        this.credentials = (await axios({
          method: 'get',
          url: this.baseUrl(),
          responseType: 'json',
        })).data;
      } catch (err) {
        this.formError = getErrorMessage(err);
      }
    },

    editCredential(credential) {
      this.formError = null;
      this.editedCredential = credential != null ? { ...credential } : {
        host: null,
        access_key_id: null,
      };
      this.editedCredentialId = credential != null ? credential.id : null;

      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }

      this.editDialog = true;
    },

    async saveCredential() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        return;
      }

      this.formSaving = true;

      try {
        await axios({
          method: this.editedCredentialId == null ? 'post' : 'put',
          url: this.editedCredentialId == null
            ? this.baseUrl()
            : `${this.baseUrl()}/${this.editedCredentialId}`,
          responseType: 'json',
          data: this.editedCredential,
        });

        this.editDialog = false;
        this.editedCredential = null;
        await this.loadCredentials();
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }
    },

    async deleteCredential(credential) {
      try {
        await axios({
          method: 'delete',
          url: `${this.baseUrl()}/${credential.id}`,
        });
        await this.loadCredentials();
      } catch (err) {
        this.formError = getErrorMessage(err);
      }
    },
  },
};
</script>
