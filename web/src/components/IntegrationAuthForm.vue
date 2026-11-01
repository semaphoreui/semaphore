<template>
  <div class="IntegrationAuthForm">
    <v-select
      v-model="authMethod"
      :label="$t('authMethod')"
      :items="authMethods"
      item-value="id"
      item-text="title"
      :disabled="disabled"
      outlined
      dense
      class="IntegrationAuthForm__method"
    ></v-select>

    <HighlightedCard v-if="authMethod">
      <template>
        <v-text-field
          v-if="HEADER_AUTH_METHODS.includes(authMethod)"
          v-model="authHeader"
          :label="$t('authHeader')"
          :disabled="disabled"
          outlined
          dense
          class="IntegrationAuthForm__header"
        ></v-text-field>

        <v-select
          v-model="authSecretId"
          :label="$t('integrationAuthSecret')"
          clearable
          :items="authSecretKeys || []"
          :loading="keys == null"
          item-value="id"
          item-text="name"
          :disabled="disabled"
          outlined
          dense
          class="IntegrationAuthForm__secret"
        ></v-select>
      </template>
    </HighlightedCard>
  </div>
</template>
<script>
import axios from 'axios';
import HighlightedCard from '@/components/HighlightedCard.vue';
import { authMethods, allowedSecretKeyTypes, HEADER_AUTH_METHODS } from '@/lib/integrationAuth';

/**
 * Webhook authorization block shared by the endpoint and the integration group forms.
 * `v-model` is the whole item; the component emits a copy with `auth_method`,
 * `auth_header` or `auth_secret_id` changed.
 */
export default {
  components: { HighlightedCard },
  props: {
    value: { type: Object, required: true },
    projectId: [Number, String],
    disabled: Boolean,
  },
  data() {
    return {
      keys: null,
      authMethods: authMethods((key) => this.$t(key)),
      HEADER_AUTH_METHODS,
    };
  },
  async created() {
    this.keys = (
      await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/keys`,
        responseType: 'json',
      })
    ).data || [];
    this.dropUnfitSecret();
  },
  watch: {
    // A key picked for a token method may not fit BasicAuth: drop it rather
    // than send a pair the server rejects.
    authMethod() {
      this.dropUnfitSecret();
    },
  },
  computed: {
    authMethod: {
      get() {
        return this.value.auth_method || '';
      },
      set(method) {
        this.update({ auth_method: method });
      },
    },
    authHeader: {
      get() {
        return this.value.auth_header || '';
      },
      set(header) {
        this.update({ auth_header: header });
      },
    },
    authSecretId: {
      get() {
        return this.value.auth_secret_id == null ? null : this.value.auth_secret_id;
      },
      set(id) {
        this.update({ auth_secret_id: id == null ? null : id });
      },
    },
    authSecretKeys() {
      if (this.keys == null) {
        return null;
      }
      const allowed = allowedSecretKeyTypes(this.authMethod);
      return this.keys.filter((key) => allowed.includes(key.type));
    },
  },
  methods: {
    update(patch) {
      this.$emit('input', { ...this.value, ...patch });
    },

    dropUnfitSecret() {
      if (this.authSecretId == null || this.authSecretKeys == null) {
        return;
      }
      if (!this.authSecretKeys.some((key) => key.id === this.authSecretId)) {
        this.update({ auth_secret_id: null });
      }
    },
  },
};
</script>
