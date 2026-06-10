<template>
  <v-skeleton-loader
    v-if="!isLoaded"
    type="table-heading, list-item-two-line, list-item-two-line, list-item-two-line"
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

    <div v-for="field in currentFields" :key="field.key">
      <v-checkbox
        v-if="field.type === 'checkbox'"
        v-model="configData[field.key]"
        :label="field.label"
        :disabled="formSaving"
        dense
      ></v-checkbox>

      <v-text-field
        v-else
        v-model="configData[field.key]"
        :label="field.label"
        :type="field.type || 'text'"
        :placeholder="field.placeholder"
        :rules="field.required ? [v => !!v || `${field.label} is required`] : []"
        :disabled="formSaving"
        outlined
        dense
      ></v-text-field>
    </div>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  props: {
    itemApp: String,
  },

  data() {
    return {
      configData: {},
      providerSchemas: {
        email: {
          name: 'Email',
          fields: [
            {
              key: 'email_sender', label: 'Sender Address', placeholder: 'noreply@example.com', required: true,
            },
            {
              key: 'email_host', label: 'SMTP Host', placeholder: 'smtp.example.com', required: true,
            },
            {
              key: 'email_port', label: 'SMTP Port', placeholder: '587', type: 'number', required: true,
            },
            { key: 'email_username', label: 'SMTP Username', required: false },
            {
              key: 'email_password', label: 'SMTP Password', type: 'password', required: false,
            },
            {
              key: 'email_secure', label: 'Secure Connection (SSL/TLS)', type: 'checkbox', required: false,
            },
            {
              key: 'email_tls', label: 'Force TLS', type: 'checkbox', required: false,
            },
            {
              key: 'email_tls_min_version', label: 'Min TLS Version', placeholder: '1.2', required: false,
            },
          ],
        },
        slack: {
          name: 'Slack',
          fields: [
            {
              key: 'slack_url', label: 'Slack Webhook URL', placeholder: 'https://hooks.slack.com/services/...', required: true,
            },
          ],
        },
      },
    };
  },

  computed: {
    isLoaded() {
      return this.item != null;
    },
    currentFields() {
      const provider = this.providerSchemas[this.itemApp];
      return provider ? provider.fields : [];
    },
  },

  watch: {
    item: {
      handler(newItem) {
        this.configData = {};
        if (newItem && newItem.config) {
          try {
            const parsedConfig = typeof newItem.config === 'string'
              ? JSON.parse(newItem.config)
              : newItem.config;

            Object.keys(parsedConfig).forEach((key) => {
              this.$set(this.configData, key, parsedConfig[key]);
            });
          } catch (e) {
            // Ignore JSON-parsing-Error
          }
        }

        if (this.currentFields) {
          this.currentFields.forEach((field) => {
            if (this.configData[field.key] === undefined) {
              this.$set(this.configData, field.key, '');
            }
          });
        }
      },
      immediate: true,
      deep: true,
    },
  },

  methods: {
    getItemsUrl() {
      return `/api/project/${this.projectId}/notifications`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/notifications/${this.itemId}`;
    },

    beforeSave() {
      this.item.type = this.itemApp;
      this.item.config = JSON.stringify(this.configData);
    },
  },
};
</script>
