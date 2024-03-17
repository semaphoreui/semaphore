<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-dialog
      v-model="paymentDialog"
      max-width="400"
      persistent
      :transition="false"
    >
      <v-card
        v-if="payment == null || payment.gateway === 'paypal' && payment.state !== 'completed'"
      >
        <v-card-title class="headline text-center">
          Replenishing your wallet
          <v-spacer></v-spacer>
          <v-btn
            @click="closePaymentDialog"
            icon
          >
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text class="text-xs-center pb-0">

          <v-slider
            v-model="currencyAmount"
            always-dirty
            :min="projectType === 'premium' ? 25 : 5"
            :max="projectType === 'premium' ? 250 : 50"
            :step="projectType === 'premium' ? 25 : 5"
            thumb-label="always"
            thumb-size="60"
            style="margin-top: 90px; margin-left: 15px; margin-right: 15px;"
          >
            <template v-slot:thumb-label="props">
              <div
                style="font-size: 20px; font-weight: bold;"
              >
                ${{ props.value }}
              </div>
            </template>
          </v-slider>

        </v-card-text>
        <v-card-actions class="pb-4 pt-0 d-block">

          <v-btn

            color="warning"
            @click="makeCoinbasePayment('coinbase')"
            large
            depressed
            style="width: 100%"
          >
            <v-icon left>mdi-bitcoin</v-icon>
            Pay Crypto
          </v-btn>

          <div class="mt-4" id="paypal-button-container"></div>

        </v-card-actions>
      </v-card>

      <v-card v-else>
        <v-card-title class="headline text-center">
          Add funds
          <v-spacer></v-spacer>
          <v-btn
            @click="closePaymentDialog"
            icon
          >
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-alert
          :value="paymentError"
          color="error"
        >
          {{ paymentError }}
        </v-alert>

        <v-card-text class="text-center">
          <div v-if="payment.state === 'completed'">
            <v-icon class="ma-2" color="success" style="font-size: 86px;">mdi-check-circle</v-icon>
            <div class="title pb-1">Payment completed</div>
            <div class="">Thank you for your payment.</div>
          </div>

          <div v-else>
            <v-progress-circular
              color="primary"
              :size="70"
              :width="7"
              indeterminate
              class="mb-3 mt-3"
            ></v-progress-circular>
            <div class="title pb-1">Awaiting payment</div>
            <div class="">Complete payment in the pop-up window.</div>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <v-tabs show-arrows class="pl-4">
      <v-tab
        v-if="projectType === 'premium'"
        key="install"
        :to="`/project/${projectId}/install`"
      >Install
      </v-tab>

      <v-tab
        v-if="projectType === ''"
        key="history"
        :to="`/project/${projectId}/history`"
      >{{ $t('history') }}
      </v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab key="settings" :to="`/project/${projectId}/settings`">{{ $t('settings') }}</v-tab>
      <v-tab
        key="billing"
        :to="`/project/${projectId}/billing`"
      >Billing
      </v-tab>
    </v-tabs>

    <v-container v-if="project != null">
      <div class="mt-7 mb-6">
        <div class="text-h3">
          <v-icon
            x-large
            color="deep-orange darken-4"
            style="margin-top: -8px;"
          >
            mdi-wallet
          </v-icon>

          ${{ project.balance }}

          <v-btn
            color="primary"
            @click="paymentDialog = true"
            icon
            x-large
            style="margin-top: -8px;"
          >
            <v-icon x-large>mdi-plus-circle</v-icon>
          </v-btn>
        </div>
      </div>

      <v-row>
        <v-col md="8" lg="8">
          <v-timeline
            align-top
            dense
            style="margin-left: -20px;"
          >
            <v-timeline-item
              fill-dot
              icon="mdi-calendar-range"
              class="text-subtitle-1 align-center"
            >
              Billing date:
              <span v-if="project.planCanceled">&mdash;</span>
              <span v-else>{{ project.planFinishDate | formatDate2 }}</span>
            </v-timeline-item>

            <v-timeline-item
              v-if="projectType === 'premium'"
              fill-dot
              icon="mdi-server"
              class="text-subtitle-1 align-center"
            >
              Servers: {{ project.servers || 0 }} / {{ plan.servers || '&mdash;' }} used
            </v-timeline-item>

            <v-timeline-item
              v-if="projectType === 'premium'"
              fill-dot
              icon="mdi-license"
              class="text-subtitle-1 align-center"
            >
              License Key: <code>{{ project.licenseKey || '&mdash;' }}</code>
              <v-btn
                v-if="project.licenseKey"
                class="ml-2"
                icon
                @click="copyToClipboard(project.licenseKey)"
              >
                <v-icon>mdi-content-copy</v-icon>
              </v-btn>
            </v-timeline-item>

            <v-timeline-item
              v-if="projectType === ''"
              fill-dot
              icon="mdi-server"
              class="text-subtitle-1 align-center"
            >Cache: {{ project.diskUsage }} / {{ plan.diskUsage }} Mb used
            </v-timeline-item>
            <v-timeline-item
              v-if="projectType === ''"
              fill-dot
              icon="mdi-cog"
              class="text-subtitle-1 align-center"
            >Runner:
              {{ Math.ceil(project.runnerUsage / 60) }} / {{ Math.ceil(plan.runnerUsage / 60) }}
              minutes used
            </v-timeline-item>
          </v-timeline>
        </v-col>
      </v-row>

      <v-row class="mt-0 mb-9" v-if="projectType === ''">
        <v-col md="4" lg="4">
          <v-card
            class="mt-4 pa-2"
            :color="$vuetify.theme.dark ? 'blue-grey darken-4' : 'grey lighten-4'"
            flat
          >
            <v-card-title class="text-h3">
              Free
              <v-spacer />
              <v-chip
                class="ml-4 mt-1 text-subtitle-1 py-3 px-4 font-weight-bold"
                v-if="project.plan === 'free'"
                color="success"
                outlined
              >Your plan
              </v-chip>

            </v-card-title>
            <v-card-text>
              <v-timeline
                align-top
                dense
                style="margin-left: -30px;"
              >
                <v-timeline-item
                  icon="mdi-server"
                  class="text-subtitle-1 align-center"
                  fill-dot
                >
                  100 Mb for cache
                </v-timeline-item>

                <v-timeline-item
                  fill-dot
                  icon="mdi-cog"
                  class="text-subtitle-1 align-center"
                >
                  50 min/mo for tasks
                </v-timeline-item>
              </v-timeline>
            </v-card-text>
            <v-card-actions>
              <v-btn large style="visibility: hidden;">
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
        <v-col md="4" lg="4">
          <v-card
            class="mt-4 pa-2"
            :color="$vuetify.theme.dark ? 'blue-grey darken-4' : 'grey lighten-4'"
            flat
          >
            <v-card-title class="text-h3">
              $5

              <v-spacer />
              <v-chip
                class="ml-4 mt-1 font-weight-bold text-center pa-4 text-subtitle-1"
                color="success"
                outlined
                :style="{'font-size': project.planCanceled ? '12px !important' : undefined}"
                v-if="project.plan === 'starter'"
              >
                Your plan
                <span class="ml-1" v-if="project.planCanceled">
                  until {{ project.planFinishDate | formatDate2 }}
                </span>
              </v-chip>
            </v-card-title>
            <v-card-text>

              <v-timeline
                align-top
                dense
                style="margin-left: -30px;"
              >
                <v-timeline-item
                  icon="mdi-server"
                  fill-dot
                  class="text-subtitle-1 align-center"
                >
                  1G for cache
                </v-timeline-item>

                <v-timeline-item
                  fill-dot
                  icon="mdi-cog"
                  class="text-subtitle-1 align-center"
                >
                  <div>1000 min/mo for tasks</div>
                </v-timeline-item>
              </v-timeline>
            </v-card-text>

            <v-card-actions>
              <v-btn
                depressed
                large
                :color="project.plan === 'free' || project.planCanceled ? 'success' : 'error'"
                style="width: 100%;"
                @click="
                project.plan === 'free' || project.planCanceled
                  ? selectPlan('starter')
                  : selectPlan('free')
              "
              >
                {{
                  project.plan === 'free'
                    ? 'Upgrade'
                    : (project.planCanceled ? 'Renew' : 'Cancel')
                }}
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
      <v-row class="mt-0 mb-9" v-else-if="projectType === 'premium'">
        <v-col v-for="plan in premiumPlans" md="4" lg="4" :key="plan.id">
          <v-card
            class="mt-4 pa-2"
            :color="$vuetify.theme.dark ? 'blue-grey darken-4' : 'grey lighten-4'"
            flat
          >
            <v-card-title class="text-h3">
              ${{ plan.price }}
              <v-spacer/>
              <v-chip
                class="text-subtitle-1 py-3 px-4 font-weight-bold"
                v-if="project.plan === plan.id"
                :style="{'font-size': project.planCanceled ? '12px !important' : undefined}"
                color="success"
                outlined
              >Your plan
                <span class="ml-1" v-if="project.planCanceled">
                  until {{ project.planFinishDate | formatDate2 }}
                </span>
              </v-chip>
            </v-card-title>

            <v-card-text>
              <v-timeline
                align-top
                dense
                style="margin-left: -30px;"
              >
                <v-timeline-item
                  icon="mdi-server"
                  fill-dot
                  class="text-subtitle-1 align-center"
                >
                  {{ plan.servers }} instance(s)
                </v-timeline-item>

                <v-timeline-item
                  fill-dot
                  icon="mdi-cog"
                  class="text-subtitle-1 align-center"
                >
                  <div>{{ plan.runners }} runners</div>
                </v-timeline-item>

                <v-timeline-item
                  fill-dot
                  icon="mdi-account-multiple"
                  class="text-subtitle-1 align-center"
                >
                  <div>{{ plan.users }} users</div>
                </v-timeline-item>
              </v-timeline>
            </v-card-text>

            <v-card-actions>
              <v-btn
                depressed
                large
                :color="project.plan !== plan.id || project.planCanceled ? 'success' : 'error'"
                style="width: 100%;"
                @click="
                project.plan !== plan.id || project.planCanceled
                  ? selectPlan(plan.id)
                  : selectPlan('free')
              "
              >
                {{ getActivePremiumButtonTitle(plan.id) }}
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>

    </v-container>
  </div>
</template>
<style lang="scss">
</style>
<script>
import EventBus from '@/event-bus';
import axios from 'axios';
import { loadScript } from '@paypal/paypal-js';
import { getErrorMessage } from '@/lib/error';

const PLANS = {
  free: {
    price: 0,
    diskUsage: 100,
    runnerUsage: 50 * 60,
  },
  starter: {
    price: 5,
    diskUsage: 1000,
    runnerUsage: 1000 * 60,
  },
  premium: {
    price: 12,
    servers: 1,
    runners: 5,
    users: 3,
  },
  premium_plus: {
    price: 50,
    servers: 3,
    runners: 20,
    users: 15,
  },
  enterprise: {
    price: 250,
    servers: 10,
    runners: 100,
    users: 70,
  },
};

export default {
  components: {},
  props: {
    projectId: Number,
    projectType: String,
  },

  data() {
    return {
      project: null,
      payment: null,
      paymentError: null,
      paymentDialog: false,
      paymentProgressTimer: null,
      currencyAmount: null,
      plan: PLANS.free,
      premiumPlans: ['premium', 'premium_plus', 'enterprise'].map((plan) => ({
        ...PLANS[plan],
        id: plan,
      })),
      paypal: null,
      paypalButton: null,
    };
  },

  watch: {
    async paymentDialog(val) {
      if (val) {
        this.currencyAmount = null;
        await this.initPaypalButton();
      } else {
        this.paypalButton.close();
      }
    },
  },

  async created() {
    await this.refreshProject();
  },

  methods: {
    async initPaypalButton() {
      if (!this.paypal) {
        try {
          this.paypal = await loadScript({ clientId: 'ATj8K7c7xzD4Z1JMozXCQGh6sUv8M9yScCeQWfVm1xaC_8UQ_0AM7IU6kr1cFbQ3FKwq0_dOz4-gkE_k' });
        } catch (error) {
          console.error('failed to load the PayPal JS SDK script', error);
        }
      }

      try {
        this.paypalButton = await this.paypal.Buttons({
          createOrder: async () => {
            try {
              this.payment = (await axios({
                method: 'post',
                url: `/billing/projects/${this.projectId}/payments`,
                responseType: 'json',
                data: {
                  currencyAmount: this.currencyAmount,
                  currency: 'usd',
                  gateway: 'paypal',
                },
              })).data;

              console.log(this.payment);

              return this.payment.gatewayTransactionId;
            } catch (err) {
              EventBus.$emit('i-snackbar', {
                color: 'error',
                text: getErrorMessage(err),
              });
              return undefined;
            }
          },

          onApprove: async () => {
            this.payment = (await axios({
              method: 'put',
              url: `/billing/projects/${this.projectId}/payments/${this.payment.number}/refresh`,
              responseType: 'json',

              data: {
                currencyAmount: this.currencyAmount,
                currency: 'usd',
                gateway: 'paypal',
              },
            })).data;

            await this.refreshProject();
          },
        });

        this.paypalButton.render('#paypal-button-container');
      } catch (error) {
        console.error('failed to render the PayPal Buttons', error);
      }
    },

    async copyToClipboard(text) {
      try {
        await window.navigator.clipboard.writeText(text);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'The license key has been copied to the clipboard.',
        });
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Can't copy the license key: ${e.message}`,
        });
      }
    },

    getActivePremiumButtonTitle(planId) {
      if (this.project.plan === 'free') {
        return 'Activate';
      }

      if (this.project.plan === planId) {
        if (this.project.planCanceled) {
          return 'Renew';
        }
        return 'Cancel';
      }

      if (PLANS[this.project.plan].price < PLANS[planId].price) {
        return 'Upgrade';
      }

      return 'Downgrade';
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    onError(e) {
      EventBus.$emit('i-snackbar', {
        color: 'error',
        text: e.message,
      });
    },

    async refreshProject() {
      this.project = (await axios({
        method: 'get',
        url: `/billing/projects/${this.projectId}`,
        responseType: 'json',
      })).data;

      this.plan = PLANS[this.project.plan];
    },

    async selectPlan(plan) {
      await this.refreshProject();

      const { price } = PLANS[plan];

      if (this.project.balance < price) {
        this.paymentDialog = true;
      } else {
        await axios({
          method: 'put',
          url: `/billing/projects/${this.projectId}`,
          responseType: 'json',
          data: {
            plan,
          },
        });
        await this.refreshProject();
      }
    },

    async refreshPayment() {
      this.payment = (await axios({
        method: 'get',
        url: `/billing/projects/${this.projectId}/payments/${this.payment.number}`,
        responseType: 'json',
      })).data;
    },

    async closePaymentDialog() {
      this.payment = null;
      this.paymentError = null;
      this.paymentDialog = false;
    },

    async makeCoinbasePayment() {
      window.gtag_report_conversion(this.currencyAmount);

      try {
        this.payment = (await axios({
          method: 'post',
          url: `/billing/projects/${this.projectId}/payments`,
          responseType: 'json',
          // headers: {
          //   authorization: `Bearer ${localStorage.getItem('authenticationToken')}`,
          // },
          data: {
            currencyAmount: this.currencyAmount,
            currency: 'usd',
            gateway: 'coinbase',
          },
        })).data;

        this.paymentError = null;
        this.paymentDialog = true;

        // eslint-disable-next-line no-promise-executor-return
        await new Promise((resolve) => setTimeout(resolve, 600));

        window.open(this.payment.hostedUrl, '_blank');

        this.paymentProgressTimer = setInterval(async () => {
          await this.refreshPayment();
          if (this.payment.state !== 'new' && this.payment.state !== 'pending') {
            clearInterval(this.paymentProgressTimer);
            await this.refreshProject();
          }
        }, 2000);
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
