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

    <div v-if="showProUser" style="margin-bottom: 30px">
      <v-alert
        class="mb-3"
        type="success"
      >
        <span>
          Congrats! You are now using a Pro subscription.
        </span>
      </v-alert>

      <div style="margin: 20px 0; font-size: 16px;">
        Are you want to make your current user <strong>Pro</strong>?
      </div>

      <div>
        <v-btn
          @click="showProUser = false"
          color="primary"
          :disabled="formSaving"
          style="width: calc(50% - 5px); margin-right: 10px;"
        >
          No
        </v-btn>
        <v-btn
          @click="makeProUser"
          color="primary"
          :disabled="formSaving"
          style="width: calc(50% - 5px);"
        >
          Yes
        </v-btn>
      </div>
    </div>

    <div v-else style=" margin-bottom: 30px;">
      <v-textarea
        class="mt-4"
        rows="4"
        auto-grow
        v-model="item.key"
        label="Subscription Key"
        :rules="[v => !!v || $t('key_required')]"
        required
        :disabled="formSaving"
        outlined
        dense
      ></v-textarea>

      <v-btn
        @click="save"
        style="width: 100%; margin-top: -5px;"
        color="primary"
        :disabled="formSaving"
      >
        <v-progress-circular
          v-if="formSaving"
          indeterminate
          color="white"
          :size="24"
        ></v-progress-circular>
        <span v-else>Activate new key</span>
      </v-btn>
    </div>

    <v-card
      v-if="item.plan"
      class="mb-3"
      style="background: var(--highlighted-card-bg-color)"
    >
      <v-card-title>Plan &amp; status</v-card-title>
      <v-card-text class="pb-2">
        <v-row>
          <v-col class="py-0">
            <v-list class="py-0" style="background: unset;">
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Plan</v-list-item-title>
                  <v-list-item-subtitle>{{ item.plan }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Expires at</v-list-item-title>
                  <v-list-item-subtitle>{{ item.expiresAt }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0" v-if="item.nodes">
                <v-list-item-content>
                  <v-list-item-title>Nodes</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.nodes }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Project runners</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.runners_used }} / {{ item.runners }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-col>
          <v-col class="py-0">
            <v-list class="py-0" style="background: unset;">
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Status</v-list-item-title>
                  <v-list-item-subtitle style="display: flex; align-items: center;">
                    <div
                      style="
                        border-radius: 100px;
                        width: 8px;
                        height: 8px;
                        background: #00bc00;
                        margin-right: 5px;
                        margin-top: 1px;
                      "
                    ></div>
                    <div>{{ item.state }}</div>
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Pro users</v-list-item-title>
                  <v-list-item-subtitle>{{ item.used }} / {{ item.users }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Terraform backends</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.terraform_backends_used }} / {{ item.terraform_backends }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-col>
        </v-row>

        <div style="
          margin-top: 20px;
          font-weight: bold;
          color: #00bc00;
        ">Renews in {{ (new Date() - new Date(item.expiresAt)) | formatMilliseconds }}</div>
      </v-card-text>
    </v-card>

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
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      tab: 0,
      showProUser: false,
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
    async makeProUser() {
      try {
        const user = (await axios.get('/api/user')).data;
        user.pro = true;
        await axios.put(`/api/users/${user.id}`, user);
        await this.loadData();
        this.$emit('save', {
          item: this.item,
          action: 'edit',
        });
        this.showProUser = false;
      } catch (err) {
        this.formError = getErrorMessage(err);
      }
    },

    async afterSave() {
      await this.loadData();
      const user = (await axios.get('/api/user')).data;
      this.showProUser = this.item.used < this.item.users && !user.pro;
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
