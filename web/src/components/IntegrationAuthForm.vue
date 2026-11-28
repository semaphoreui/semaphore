<template>
  <div class="IntegrationAuthForm" data-testid="integrationAuthForm">
    <v-select
      v-model="authMethod"
      :label="$t('authMethod')"
      :items="authMethods"
      item-value="id"
      item-title="title"
      :disabled="disabled"
      variant="outlined"
      density="compact"
      class="IntegrationAuthForm__method"
      data-testid="integrationAuthForm-authMethod"
    ></v-select>

    <HighlightedCard v-if="authMethod" data-testid="integrationAuthForm-auth">
      <v-text-field
        v-if="HEADER_AUTH_METHODS.includes(authMethod)"
        v-model="authHeader"
        :label="$t('authHeader')"
        :disabled="disabled"
        variant="outlined"
        density="compact"
        class="IntegrationAuthForm__header"
        data-testid="integrationAuthForm-authHeader"
      ></v-text-field>

      <v-select
        v-model="authSecretId"
        :label="$t('integrationAuthSecret')"
        clearable
        :items="authSecretKeys || []"
        :loading="keys == null"
        item-value="id"
        item-title="name"
        :disabled="disabled"
        variant="outlined"
        density="compact"
        class="IntegrationAuthForm__secret"
        data-testid="integrationAuthForm-authSecretId"
      ></v-select>
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
    modelValue: { type: Object, required: true },
    projectId: [Number, String],
    disabled: Boolean,
  },
  emits: ['update:modelValue'],
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
        return this.modelValue.auth_method || '';
      },
      set(method) {
        this.update({ auth_method: method });
      },
    },
    authHeader: {
      get() {
        return this.modelValue.auth_header || '';
      },
      set(header) {
        this.update({ auth_header: header });
      },
    },
    authSecretId: {
      get() {
        return this.modelValue.auth_secret_id == null ? null : this.modelValue.auth_secret_id;
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
      this.$emit('update:modelValue', { ...this.modelValue, ...patch });
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
