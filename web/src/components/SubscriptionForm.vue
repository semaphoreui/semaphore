<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="item != null"
    data-testid="subscription-form"
    :data-state="formSaving ? 'loading' : formError ? 'error' : 'ready'"
  >
    <v-alert
      :model-value="!!formError"
      :type="(formError || '').includes('already activated') ? 'warning' : 'error'"
      data-testid="subscription-error"
      >{{ formError }}
    </v-alert>

    <div v-if="showProUser" style="margin-bottom: 30px" data-testid="subscription-proUser">
      <v-alert class="mb-3" type="success" data-testid="subscription-proUserCongrats">
        <span> Congrats! You are now using a Pro subscription. </span>
      </v-alert>

      <div style="margin: 20px 0; font-size: 16px">
        Are you want to make your current user <strong>Pro</strong>?
      </div>

      <div>
        <v-btn
          @click="showProUser = false"
          data-testid="subscription-proUserNo"
          color="primary"
          :disabled="formSaving"
          style="width: calc(50% - 5px); margin-right: 10px"
        >
          No
        </v-btn>
        <v-btn
          @click="makeProUser"
          data-testid="subscription-proUserYes"
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
        data-testid="subscription-status-active"
      >
        You {{ item.plan.startsWith('enterprise_') ? 'Enterprise' : 'PRO' }} subscription is active.
      </div>

      <div
        v-else-if="item.state === 'expired'"
        style="line-height: 1.3; font-weight: bold; color: rgb(188, 0, 0)"
        class="mb-5"
        data-testid="subscription-status-expired"
      >
        You {{ item.plan.startsWith('enterprise_') ? 'Enterprise' : 'PRO' }}
        subscription has expired.
      </div>

      <div v-else style="line-height: 1.3" data-testid="subscription-status-none">
        Enter your subscription key to unlock advanced features, or get a new one instantly.
      </div>

      <v-textarea
        class="SubscriptionForm__key mt-4"
        rows="4"
        auto-grow
        v-model="item.key"
        label="Enter your PRO or EE key"
        :rules="[(v) => !!v || $t('key_required')]"
        required
        :disabled="formSaving || item.managed_by_config"
        variant="outlined"
        density="compact"
        data-testid="subscription-key"
      ></v-textarea>

      <v-menu location="bottom" v-if="item.state === 'active' || item.state === 'expired'">
        <template v-slot:activator="{ props }">
          <v-btn
            color="primary"
            v-bind="props"
            icon
            data-testid="subscription-actionsMenu"
            size="40"
            class="SubscriptionForm__fab"
            style="position: absolute; top: 15px; right: -15px"
          >
            <v-icon>mdi-dots-horizontal</v-icon>
          </v-btn>
        </template>

        <v-list>
          <v-list-item link @click="reloadToken" data-testid="subscription-actionsMenu-reload">
            <template v-slot:prepend>
              <v-icon>mdi-refresh</v-icon>
            </template>
            <v-list-item-title>Reload</v-list-item-title>
          </v-list-item>
          <v-list-item link @click="uploadKeyFile" data-testid="subscription-actionsMenu-upload">
            <template v-slot:prepend>
              <v-icon>mdi-upload</v-icon>
            </template>
            <v-list-item-title>Upload</v-list-item-title>
          </v-list-item>
          <v-list-item link @click="resetToken" data-testid="subscription-actionsMenu-reset">
            <template v-slot:prepend>
              <v-icon>mdi-delete</v-icon>
            </template>
            <v-list-item-title>Reset</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-btn
        v-else
        color="primary"
        icon
        size="40"
        class="SubscriptionForm__fab"
        style="position: absolute; top: 30px; right: -15px"
        @click="uploadKeyFile()"
        data-testid="subscription-upload"
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
            data-testid="subscription-activate"
          >
            <v-progress-circular
              v-if="formSaving"
              indeterminate
              color="white"
              bg-color="transparent"
              :size="24"
              data-testid="subscription-activateSaving"
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
            data-testid="subscription-buy"
            :href="`https://portal.semaphoreui.com/buy_pro?utm_source=app&utm_content=feature_${feature}`"
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
        variant="outlined"
        data-testid="subscription-trial"
        :href="`https://portal.semaphoreui.com/start_trial?utm_source=app&utm_content=feature_${feature}`"
      >
        Get 30-day free trial
      </v-btn>
    </div>

    <v-card
      v-if="item.plan"
      class="SubscriptionForm__plan mb-3"
      style="background: var(--highlighted-card-bg-color)"
      data-testid="subscription-plan"
    >
      <v-card-title>Plan &amp; status</v-card-title>
      <v-card-text class="pb-2">
        <v-list class="py-0 pb-5" style="background: unset" v-if="item.company">
          <v-list-item class="pa-0">
            <v-list-item-title>Subscription holder</v-list-item-title>
            <v-list-item-subtitle
              data-testid="subscription-planHolder"
            >{{ item.company }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>

        <v-row class="SubscriptionForm__planRow">
          <v-col class="py-0">
            <v-list class="py-0" style="background: unset">
              <v-list-item class="px-0 py-3">
                <v-list-item-title>Plan</v-list-item-title>
                <v-list-item-subtitle
                  data-testid="subscription-planName"
                >{{ item.plan }}</v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3">
                <v-list-item-title>Expires at</v-list-item-title>
                <v-list-item-subtitle
                  data-testid="subscription-planExpiresAt"
                >{{ item.expiresAt }}</v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3" v-if="item.nodes">
                <v-list-item-title>Nodes</v-list-item-title>
                <v-list-item-subtitle data-testid="subscription-planNodes">
                  {{ item.nodes_used }} / {{ item.nodes }}
                </v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3" v-if="item.runners < 100000">
                <v-list-item-title>Project runners</v-list-item-title>
                <v-list-item-subtitle data-testid="subscription-planRunners">
                  {{ item.runners_used }} / {{ item.runners }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-col>
          <v-col class="py-0">
            <v-list class="py-0" style="background: unset">
              <v-list-item class="px-0 py-3">
                <v-list-item-title>Status</v-list-item-title>
                <v-list-item-subtitle
                  style="display: flex; align-items: center"
                  data-testid="subscription-planStatus"
                >
                  <div
                    style="
                      border-radius: 100px;
                      width: 8px;
                      height: 8px;
                      margin-right: 5px;
                      margin-top: 1px;
                    "
                    :style="{
                      backgroundColor: item.state === 'active' ? '#00bc00' : '#ff0000',
                    }"
                  ></div>
                  <div>{{ item.state }}</div>
                </v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3" v-if="item.users < 100000">
                <v-list-item-title>Pro users</v-list-item-title>
                <v-list-item-subtitle
                  data-testid="subscription-planUsers"
                >{{ item.used }} / {{ item.users }}</v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3" v-if="item.terraform_states < 100000">
                <v-list-item-title>Terraform backends</v-list-item-title>
                <v-list-item-subtitle data-testid="subscription-planTerraformStates">
                  {{ item.terraform_states_used }} / {{ item.terraform_states }}
                </v-list-item-subtitle>
              </v-list-item>
              <v-list-item class="px-0 py-3" v-if="item.uis">
                <v-list-item-title>UIs</v-list-item-title>
                <v-list-item-subtitle data-testid="subscription-planUis">
                  {{ item.uis_used }} / {{ item.uis }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-col>
        </v-row>

        <div
          v-if="subscriptionTimeRemainingMs > 0"
          style="margin-top: 20px; font-weight: bold; color: #00bc00"
          data-testid="subscription-renews"
        >
          Renews in {{ $filters.formatMilliseconds(subscriptionTimeRemainingMs) }}
          <span>(if auto-renew is activated)</span>
        </div>
      </v-card-text>
    </v-card>

    <div v-else class="mb-4 mt-2" data-testid="subscription-help">
      <div>
        Need help?
        <a
          target="_blank"
          class="LinkHoverable"
          href="https://semaphoreui.com/contact/"
          data-testid="subscription-contactSupport"
        >
          Contact support
          <v-icon
            size="small"
            color="primary"
            style="transform: translateY(-1px)"
          >mdi-open-in-new</v-icon>
        </a>
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

// The key field: the dense outlined auto-grow textarea of Vuetify 2 (`.v-textarea textarea`) —
// lines of 28 px, 6 px above the text, none below, at least rows × 24 px high (its `rowHeight`):
// a 102 px box for four rows, not 112 px.
.SubscriptionForm__key.v-textarea--field-sizing textarea.v-field__input {
  line-height: 28px;
  padding-top: 6px;
  padding-bottom: 0;
  min-height: calc(4 * 24px + 6px);
}

// The round buttons at the top right of the key field: `fab small` of Vuetify 2 — 40 px
// (`size="40"`), shadow of elevation 6 (`.v-btn--fab`; Vuetify 4 has elevation classes up to 5).
.v-btn.SubscriptionForm__fab {
  box-shadow: 0 3px 5px -1px rgba(0, 0, 0, .2), 0 6px 10px 0 rgba(0, 0, 0, .14),
    0 1px 18px 0 rgba(0, 0, 0, .12);
}

// The lines of the plan card as `v-list-item-content` of Vuetify 2 laid them out: the title 2 px
// above the subtitle (`.v-list-item__content > *:not(:last-child)`), both on one line
// (`white-space: nowrap`).
.SubscriptionForm__plan .v-list-item-title {
  margin-bottom: 2px;
}

// The two columns of the plan: the gutters of the Vuetify 2 grid (`.row` margin -12 px, the
// columns `py-0`), so the lists start 12 px higher and nothing is between the columns when they
// wrap; a column is as wide as its longest line (`.col`: flex-basis 0, min-width auto), so on a
// phone the columns go one under the other instead of cutting the values with an ellipsis.
.SubscriptionForm__planRow {
  margin-block: -12px;
  row-gap: 0;

  .v-list-item__content {
    min-width: auto;
    overflow: visible;
  }

  .v-list-item-subtitle {
    white-space: nowrap;
  }
}
</style>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  props: {
    feature: String,
  },

  emits: ['save'],

  data() {
    return {
      tab: 0,
      showProUser: false,
    };
  },

  computed: {
    subscriptionTimeRemainingMs() {
      const ret = new Date(this.item.expiresAt) - new Date();
      return ret;
    },

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
