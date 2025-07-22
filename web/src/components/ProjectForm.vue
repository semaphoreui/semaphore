<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('projectName')"
      :rules="[v => !!v || $t('project_name_required')]"
      required
      :disabled="formSaving"
      data-testid="newProject-name"
      outlined
      dense
    ></v-text-field>

    <v-text-field
      v-model.number="item.max_parallel_tasks"
      :label="$t('maxNumberOfParallelTasksOptional')"
      :disabled="formSaving"
      :rules="[
        v => (v == null || v === '' || Math.floor(v) === v) || $t('mustBeInteger'),
        v => (v == null || v === '' || v >= 0) || $t('mustBe0OrGreater'),
      ]"
      hint="Should be 0 or greater, 0 - unlimited."
      type="number"
      :step="1"
      outlined
      dense
    ></v-text-field>

    <v-text-field
      v-model="item.alert_chat"
      :label="$t('telegramChatIdOptional')"
      :disabled="formSaving"
      data-testid="newProject-tg"
      outlined
      dense
    ></v-text-field>

    <v-checkbox
      class="mt-0"
      v-model="item.alert"
      :label="$t('allowAlertsForThisProject')"
      data-testid="newProject-alert"
    ></v-checkbox>

    <v-row align="center">
      <v-col class="shrink">

        <v-btn
          color="blue-grey"
          @click="sendTestNotification()"
          :disabled="testNotificationProgress"
          min-width="170"
          data-testid="settings-testNotificationProgress"
        >{{ $t('backup') }}
        </v-btn>

        <v-progress-linear
          v-if="testNotificationProgress"
          color="primary accent-4"
          indeterminate
          rounded
          height="36"
          style="margin-top: -36px"
        ></v-progress-linear>

      </v-col>
      <v-col class="grow">
        <div style="font-size: 14px;">
          {{ $t('downloadTheProjectBackupFile') }}
        </div>
      </v-col>
    </v-row>
    <v-btn v-if="itemId !== 'new'" color="blue-grey">
      Test notifications
    </v-btn>

    <v-switch
      v-if="itemId === 'new'"
      v-model="item.demo"
      label="Demo"
      style="position: absolute; left: 24px; bottom: 15px;"
      hide-details
    />

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  props: {
  },
  data() {
    return {
      testNotificationProgress: false,
    };
  },
  methods: {
    sendTestNotification() {
      this.testNotificationProgress = true;
      try {
        // TODO: Implement the actual notification sending logic

      } finally {
        this.testNotificationProgress = false;
      }
    },
    getItemsUrl() {
      return '/api/projects';
    },
    getSingleItemUrl() {
      return `/api/project/${this.itemId}`;
    },
    beforeSave() {
      if (this.item.max_parallel_tasks === '') {
        this.item.max_parallel_tasks = 0;
      }
    },
  },
};
</script>
