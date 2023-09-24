<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-dialog
      v-model="paymentDialog"
      max-width="400"
      persistent
      :transition="false"
    >
      <v-card v-if="payment == null">
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
        <v-card-text class="text-xs-center pb-0">
          lease make a payment

          <v-slider
            v-model="currencyAmount"
            always-dirty
            :min="5"
            :max="50"
            :step="5"
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
        <v-card-actions class="pb-4 pt-0">

          <v-btn
            color="warning"
            @click="makePayment('coinbase')"
            large
            style="width: 100%"
          >
            <v-icon left>mdi-bitcoin</v-icon>
            Pay Crypto
          </v-btn>

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
            <v-icon class="ma-2" color="success" style="font-size: 86px;">check_circle</v-icon>
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
      <v-tab key="history" :to="`/project/${projectId}/history`">{{ $t('history') }}</v-tab>
      <v-tab key="activity" :to="`/project/${projectId}/activity`">{{ $t('activity') }}</v-tab>
      <v-tab key="settings" :to="`/project/${projectId}/settings`">{{ $t('settings') }}</v-tab>
      <v-tab
        key="billing"
        :to="`/project/${projectId}/billing`"
      >Billing</v-tab>
    </v-tabs>

    <v-container v-if="project != null">
      <div>
<!--        <v-icon>mdi-wallet</v-icon>-->
        Your balance: ${{ project.balance }}
        <v-btn
          color="primary"
          @click="paymentDialog = true"
        >
          <v-icon left>mdi-plus-circle</v-icon>
          Add fonds
        </v-btn>
      </div>
      <hr />
      <div>
        Billing date: {{ project.planFinishDate }}
      </div>
      <div>
        Cache used: {{ project.diskUsage }} from 100 Mb
      </div>
      <div>
        Runner used: {{ project.runnerUsage }} from 1000 minutes
      </div>
      <hr />

      <v-row no-gutters>
        <v-col md="4" lg="4">
          <h2>Free</h2>
          <div>100 Mb for cache</div>
          <div>50 minutes/month</div>
          <div>
            <span v-if="project.plan === 'free'">Your plan</span>
          </div>
        </v-col>
        <v-col md="4" lg="4">
          <h2>$5</h2>
          <div>1G for cache</div>
          <div>1000 minutes/month</div>
          <div>
            <span v-if="project.plan === 'starter'">Your plan</span>
          </div>
          <div>
            <v-btn
              color="primary"
              @click="
                project.plan === 'free' || project.planCanceled
                  ? selectPlan('starter')
                  : selectPlan('free')
              "
            >
              {{
                project.plan === 'free'
                  ? 'Select'
                  : (project.planCanceled ? 'Continue' : 'Cancel')
              }}
            </v-btn>
          </div>
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
import { getErrorMessage } from '@/lib/error';

const PLAN_PRICES = {
  free: 0,
  starter: 5,
};

export default {
  components: {},
  props: {
    projectId: Number,
  },

  data() {
    return {
      project: null,
      payment: null,
      paymentError: null,
      paymentDialog: false,
      paymentProgressTimer: null,
      currencyAmount: null,
    };
  },

  async created() {
    await this.refreshProject();
  },

  methods: {
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
    },

    async selectPlan(plan) {
      await this.refreshProject();

      const price = PLAN_PRICES[plan];

      if (this.project.plan === 'free' && this.project.balance < price) {
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

    async makePayment() {
      try {
        this.payment = (await axios({
          method: 'post',
          url: `/billing/projects/${this.projectId}/payments`,
          responseType: 'json',
          headers: {
            authorization: `Bearer ${localStorage.getItem('authenticationToken')}`,
          },
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
