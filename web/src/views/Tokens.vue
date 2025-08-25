<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat>
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('api_tokens') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <a :href="`${this.systemInfo?.web_host || ''}/swagger/index.html`" class="mr-6">
        {{ $t('API Reference') }}
      </a>

      <v-btn
        color="primary"
        @click="newToken()"
      >{{ $t('New Token') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >

      <template v-slot:item.id="{ item }">

        <code v-if="item.token_id && item.show_token_id" class="mr-2">{{ item.token_id }}</code>
        <code v-else class="mr-2">{{ item.id }}***</code>

        <v-btn
          icon
          v-if="item.token_id && !item.show_token_id"
          @click="showToken(item.id)"
        >
          <v-icon>mdi-eye</v-icon>
        </v-btn>

        <CopyClipboardButton
          v-if="item.token_id"
          :text="item.token_id"
          success-message="The token has been copied to the clipboard."
        />

      </template>

      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate}}
      </template>

      <template v-slot:item.expired="{ item }">
        <div class="pr-4">
          <v-chip v-if="item.expired" style="font-weight: bold;" color="error">
            Expired
          </v-chip>
          <v-chip v-else style="font-weight: bold;" color="success">
            Active
          </v-chip>
        </div>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="deleteItem(item.id)"
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
  },

  data() {
    return {
      newRunnerTokenDialog: null,
    };
  },

  methods: {

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

    async newToken() {
      const res = (await axios({
        method: 'post',
        url: '/api/user/tokens',
        responseType: 'json',
        data: {},
      })).data;
      await this.loadItems();

      const i = this.items.findIndex((item) => res.id.startsWith(item.id));
      if (i === -1) {
        return;
      }

      this.items.splice(i, 1, {
        ...this.items[i],
        token_id: res.id,
      });
    },

    getHeaders() {
      return [{
        text: this.$i18n.t('token'),
        value: 'id',
      }, {
        text: this.$i18n.t('created'),
        value: 'created',
      }, {
        text: this.$i18n.t('status'),
        value: 'expired',
      }, {
        text: '',
        value: 'actions',
        sortable: false,
        width: 70,
      }];
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
