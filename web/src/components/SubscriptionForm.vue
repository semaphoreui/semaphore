<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      :type="(formError || '').includes('already activated') ? 'warning' : 'error'"
    >{{ formError }}
    </v-alert>

    <v-textarea
      class="mt-4"
      rows="4"
      v-model="item.key"
      label="Subscription Key"
      :rules="[v => !!v || $t('key_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-textarea>

    <v-row v-if="item.plan" class="mb-3">
      <v-col class="py-0">
        <v-list class="py-0">
          <v-list-item class="pa-0">
            <v-list-item-title>Plan</v-list-item-title>
            <v-list-item-subtitle>{{ item.plan }}</v-list-item-subtitle>
          </v-list-item>
          <v-list-item class="pa-0">
            <v-list-item-title>Pro users</v-list-item-title>
            <v-list-item-subtitle>{{ item.users }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col class="py-0">
        <v-list class="py-0">
          <v-list-item class="pa-0">
            <v-list-item-title>Expires at</v-list-item-title>
            <v-list-item-subtitle>{{ item.expiresAt }}</v-list-item-subtitle>
          </v-list-item>
          <v-list-item class="pa-0">
            <v-list-item-title>Status</v-list-item-title>
            <v-list-item-subtitle>
              <v-chip :color="statusColor" label class="mt-1">{{ item.state }}</v-chip>
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>

    <div v-else class="mb-4 mt-2">

      <div>
        Don't have subscription key? <a
        target="_blank"
        href="https://portal.semaphoreui.com/auth/login?new_project=premium"
      >Get one</a>.
      </div>
    </div>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      tab: 0,
    };
  },

  computed: {
    isNew() {
      return false;
    },

    statusColor() {
      switch (this.item.state) {
        case 'expired':
          return 'error';
        case 'active':
          return 'success';
        default:
          return '';
      }
    },
  },

  methods: {
    async afterSave() {
      await this.loadData();
    },

    getItemsUrl() {
      return '/api/subscription';
    },

    getSingleItemUrl() {
      return '/api/subscription';
    },

    getRequestOptions() {
      return {
        method: 'post',
        url: '/api/subscription',
      };
    },
  },
};
</script>
