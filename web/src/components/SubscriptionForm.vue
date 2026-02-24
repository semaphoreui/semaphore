<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
    <v-alert
      :value="formError"
      :type="(formError || '').includes('already activated') ? 'warning' : 'error'"
      >{{ formError }}
    </v-alert>

    <div v-if="showProUser" style="margin-bottom: 30px">
      <v-alert class="mb-3" type="success">
        <span> Congrats! You are now using a Pro subscription. </span>
      </v-alert>

      <div style="margin: 20px 0; font-size: 16px">
        Are you want to make your current user <strong>Pro</strong>?
      </div>

      <div>
        <v-btn
          @click="showProUser = false"
          color="primary"
          :disabled="formSaving"
          style="width: calc(50% - 5px); margin-right: 10px"
        >
          No
        </v-btn>
        <v-btn
          @click="makeProUser"
          color="primary"
          :disabled="formSaving"
          style="width: calc(50% - 5px)"
        >
          Yes
        </v-btn>
      </div>
    </div>

    <div v-else style="margin-bottom: 30px; position: relative">
      <div
        v-if="item.state === 'active'"
        style="line-height: 1.3; font-weight: bold; color: rgb(0, 188, 0)"
        class="mb-5"
      >
        You {{ item.plan.startsWith('enterprise_') ? 'Enterprise' : 'PRO' }} subscription is active.
      </div>
      <div v-else style="line-height: 1.3">
        Enter your subscription key to unlock advanced features, or get a new one instantly.
      </div>

      <v-textarea
        class="mt-4"
        rows="4"
        auto-grow
        v-model="item.key"
        label="Enter your PRO or EE key"
        :rules="[(v) => !!v || $t('key_required')]"
        required
        :disabled="formSaving || item.managed_by_config"
        outlined
        dense
      ></v-textarea>

      <v-menu offset-y v-if="item.state === 'active'">
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            color="primary"
            v-bind="attrs"
            v-on="on"
            fab
            small
            style="position: absolute; top: 15px; right: -15px"
          >
            <v-icon>mdi-dots-horizontal</v-icon>
          </v-btn>
        </template>

        <v-list>
          <v-list-item link @click="reloadToken">
            <v-list-item-icon>
              <v-icon>mdi-refresh</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Reload</v-list-item-title>
          </v-list-item>
          <v-list-item link @click="uploadKeyFile">
            <v-list-item-icon>
              <v-icon>mdi-upload</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Upload</v-list-item-title>
          </v-list-item>
          <v-list-item link @click="resetToken">
            <v-list-item-icon>
              <v-icon>mdi-delete</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Reset</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-btn
        v-else
        color="primary"
        v-bind="attrs"
        v-on="on"
        fab
        small
        style="position: absolute; top: 30px; right: -15px"
        @click="uploadKeyFile()"
      >
        <v-icon>mdi-upload</v-icon>
      </v-btn>

      <v-row>
        <v-col>
          <v-btn
            @click="save"
            style="width: 100%"
            color="success"
            :disabled="formSaving || item.managed_by_config"
          >
            <v-progress-circular
              v-if="formSaving"
              indeterminate
              color="white"
              :size="24"
            ></v-progress-circular>
            <span v-else>Activate New Key</span>
          </v-btn>
        </v-col>
        <v-col>
          <v-btn
            style="width: 100%"
            color="primary"
            :disabled="formSaving"
            target="_blank"
            href="https://portal.semaphoreui.com/buy_pro?utm_source=app"
            >Buy Pro</v-btn
          >
        </v-col>
      </v-row>

      <v-btn
        v-if="item.state !== 'active'"
        style="width: 100%"
        color="primary"
        class="mt-4"
        :disabled="formSaving"
        target="_blank"
        outlined
        href="https://portal.semaphoreui.com/start_trial?utm_source=app"
      >
        Get 30-day free trial
      </v-btn>
    </div>

    <v-card v-if="item.plan" class="mb-3" style="background: var(--highlighted-card-bg-color)">
      <v-card-title>Plan &amp; status</v-card-title>
      <v-card-text class="pb-2">
        <v-list class="py-0 pb-5" style="background: unset" v-if="item.company">
          <v-list-item class="pa-0">
            <v-list-item-content class="py-0">
              <v-list-item-title>Subscription holder</v-list-item-title>
              <v-list-item-subtitle>{{ item.company }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>

        <v-row>
          <v-col class="py-0">
            <v-list class="py-0" style="background: unset">
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
                    {{ item.nodes_used }} / {{ item.nodes }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0" v-if="item.runners < 100000">
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
            <v-list class="py-0" style="background: unset">
              <v-list-item class="pa-0">
                <v-list-item-content>
                  <v-list-item-title>Status</v-list-item-title>
                  <v-list-item-subtitle style="display: flex; align-items: center">
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
              <v-list-item class="pa-0" v-if="item.users < 100000">
                <v-list-item-content>
                  <v-list-item-title>Pro users</v-list-item-title>
                  <v-list-item-subtitle>{{ item.used }} / {{ item.users }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0" v-if="item.terraform_states < 100000">
                <v-list-item-content>
                  <v-list-item-title>Terraform backends</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.terraform_states_used }} / {{ item.terraform_states }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
              <v-list-item class="pa-0" v-if="item.uis">
                <v-list-item-content>
                  <v-list-item-title>UIs</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.uis_used }} / {{ item.uis }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-col>
        </v-row>

        <div style="margin-top: 20px; font-weight: bold; color: #00bc00">
          Renews in {{ (new Date() - new Date(item.expiresAt)) | formatMilliseconds }}
          <span>(if auto-renew is activated)</span>
        </div>
      </v-card-text>
    </v-card>

    <div v-else class="mb-4 mt-2">
      <div>
        Need help?
        <a
          target="_blank"
          class="LinkHoverable"
          href="https://portal.semaphoreui.com/auth/login?new_project=premium"
          >Contact support</a
        >
      </div>
    </div>
  </v-form>
</template>
<style lang="scss">
.LinkHoverable {
  text-decoration: none;
  &:hover {
    text-decoration: underline;
  }
}
</style>
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
    uploadKeyFile() {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = '.txt,.key,.pem,.lic';
      input.onchange = (e) => {
        const file = e.target.files[0];
        if (file) {
          const reader = new FileReader();
          reader.onload = (event) => {
            this.item.key = event.target.result.trim();
          };
          reader.readAsText(file);
        }
      };
      input.click();
    },

    async resetToken() {
      this.formError = null;
      this.formSaving = true;
      try {
        await axios.delete('/api/subscription');
        await this.loadData();
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }
    },
    async reloadToken() {
      this.formError = null;
      this.formSaving = true;
      try {
        await axios.post('/api/subscription/refresh');
        await this.loadData();
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }
    },
    afterLoadData() {
      if (this.item.error) {
        this.formError = this.item.error;
      }
    },

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
