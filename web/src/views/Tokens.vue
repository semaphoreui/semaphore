<template>
  <div
    v-if="items != null"
    data-testid="tokens-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <v-dialog v-model="newTokenDialog" max-width="480" persistent>
      <v-card
        data-testid="tokensNewDialog"
        :data-state="creatingToken ? 'loading' : 'ready'"
      >
        <v-card-title data-testid="tokensNewDialog-title">{{ $t('New Token') }}</v-card-title>
        <v-card-text class="pb-0 pt-4">
          <v-text-field
            v-model="newTokenName"
            :label="$t('tokenName')"
            variant="outlined"
            density="compact"
            data-testid="token-name"
          />

          <v-select
            v-model="expiresInDays"
            :items="expiryOptions"
            item-title="text"
            :label="$t('Expires')"
            variant="outlined"
            density="compact"
            data-testid="token-expires"
          />
          <v-text-field
            v-if="expiresInDays === 'custom'"
            v-model="customExpiresAt"
            type="datetime-local"
            :label="$t('Expiration date')"
            variant="outlined"
            density="compact"
            data-testid="token-expiresAt"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            color="primary"
            @click="newTokenDialog = false"
            :disabled="creatingToken"
            data-testid="tokensNewDialog-cancel"
          >
            {{ $t('Cancel') }}
          </v-btn>
          <v-btn
            variant="text"
            color="primary"
            @click="createToken()"
            :loading="creatingToken"
            data-testid="tokensNewDialog-create"
          >
            {{ $t('Create') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar flat>
      <v-btn icon class="mr-4" @click="returnToProjects()" data-testid="tokens-back">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="tokens-title"
      >{{ $t('api_tokens') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <a
        :href="`${systemInfo?.web_host || ''}/swagger/index.html`"
        class="mr-6"
        target="_blank"
        data-testid="tokens-apiReference"
      >
        {{ $t('API Reference') }}
      </a>

      <v-btn
        color="primary"
        variant="elevated"
        @click="newToken()"
        data-testid="tokens-new"
      >{{ $t('New Token') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page-options="[20]"
      :items-per-page="20"
      :mobile-breakpoint="600"
      data-testid="tokens-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.id="{ item }">
        <code
          v-if="item.token_id && item.show_token_id"
          class="mr-2"
          data-testid="tokens-rowToken"
        >{{ item.token_id }}</code>
        <code v-else class="mr-2" data-testid="tokens-rowId">{{ item.id }}***</code>

        <v-btn
          icon
          variant="text"
          v-if="item.token_id && !item.show_token_id"
          @click="showToken(item.id)"
          data-testid="tokens-rowShow"
        >
          <v-icon>mdi-eye</v-icon>
        </v-btn>

        <CopyClipboardButton
          v-if="item.token_id"
          :text="item.token_id"
          success-message="The token has been copied to the clipboard."
          test-id="tokens-rowCopy"
        />
      </template>

      <template v-slot:item.name="{ item }">
        {{ item.name || '—' }}
      </template>

      <template v-slot:item.created="{ item }">
        {{ $filters.formatDate(item.created) }}
      </template>

      <template v-slot:item.expires_at="{ item }">
        <span
          v-if="item.expires_at"
          data-testid="tokens-rowExpiresAt"
        >{{ $filters.formatDate(item.expires_at) }}</span>
        <span
          v-else
          class="text-disabled"
          data-testid="tokens-rowExpiresNever"
        >{{ $t('Never') }}</span>
      </template>

      <template v-slot:item.expired="{ item }">
        <div class="pr-4">
          <v-chip
            v-if="item.expired"
            style="font-weight: bold"
            variant="flat"
            color="error"
            data-testid="tokens-rowStatus-revoked"
          >
            {{ $t('Revoked') }}
          </v-chip>
          <v-chip
            v-else-if="isExpiredByTime(item)"
            style="font-weight: bold"
            variant="flat"
            color="warning"
            data-testid="tokens-rowStatus-expired"
          >
            {{ $t('Expired') }}
          </v-chip>
          <v-chip
            v-else
            style="font-weight: bold"
            variant="flat"
            color="success"
            data-testid="tokens-rowStatus-active"
          >
            {{ $t('Active') }}
          </v-chip>
        </div>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="tokens-rowActions">
          <v-btn
            icon
            variant="text"
            class="mr-1"
            @click="deleteItem(item.id)"
            data-testid="tokens-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import EventBus from '@/event-bus';
import ItemListPageBase from '@/components/ItemListPageBase';
import axios from 'axios';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  mixins: [ItemListPageBase],

  components: {
    CopyClipboardButton,
  },

  props: {
    systemInfo: Object,
  },

  computed: {
    expiryOptions() {
      return [
        { text: this.$t('Never'), value: 0 },
        { text: '30 days', value: 30 },
        { text: '90 days', value: 90 },
        { text: '1 year', value: 365 },
        { text: this.$t('Custom'), value: 'custom' },
      ];
    },
  },

  data() {
    return {
      newTokenName: '',
      newTokenDialog: false,
      creatingToken: false,
      expiresInDays: 0,
      customExpiresAt: null,
    };
  },

  methods: {
    isExpiredByTime(item) {
      if (!item.expires_at) return false;
      return new Date(item.expires_at).getTime() <= Date.now();
    },

    async showToken(token) {
      const i = this.items.findIndex((item) => item.id === token);
      if (i === -1) {
        return;
      }

      this.items.splice(i, 1, {
        ...this.items[i],
        show_token_id: true,
      });
    },

    newToken() {
      this.newTokenName = '';
      this.expiresInDays = 0;
      this.customExpiresAt = null;
      this.newTokenDialog = true;
    },

    async createToken() {
      let expiresAt = null;
      if (this.expiresInDays === 'custom') {
        if (!this.customExpiresAt) {
          return;
        }
        expiresAt = new Date(this.customExpiresAt).toISOString();
      } else if (this.expiresInDays > 0) {
        const d = new Date();
        d.setDate(d.getDate() + this.expiresInDays);
        expiresAt = d.toISOString();
      }

      this.creatingToken = true;
      try {
        const res = (
          await axios({
            method: 'post',
            url: '/api/user/tokens',
            responseType: 'json',
            data: {
              expires_at: expiresAt,
              name: this.newTokenName,
            },
          })
        ).data;
        await this.loadItems();

        const i = this.items.findIndex((item) => res.id.startsWith(item.id));
        if (i !== -1) {
          this.items.splice(i, 1, {
            ...this.items[i],
            token_id: res.id,
          });
        }
        this.newTokenDialog = false;
      } finally {
        this.creatingToken = false;
      }
    },

    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
        },
        {
          title: this.$t('token'),
          key: 'id',
        },
        {
          title: this.$t('created'),
          key: 'created',
        },
        {
          title: this.$t('Expires'),
          key: 'expires_at',
        },
        {
          title: this.$t('status'),
          key: 'expired',
        },
        {
          title: '',
          key: 'actions',
          sortable: false,
          width: 70,
        },
      ];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/user/tokens';
    },

    getSingleItemUrl() {
      return `/api/user/tokens/${this.itemId}`;
    },

    getEventName() {
      return 'i-token';
    },
  },
};
</script>
