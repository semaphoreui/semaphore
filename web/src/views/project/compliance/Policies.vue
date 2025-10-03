<template>
  <div class="policies-container">
    <!-- Header -->
    <div class="d-flex justify-space-between align-center mb-6">
      <div>
        <h2 class="text-h5 mb-2">{{ $t('compliance.policies') }}</h2>
        <p class="text-body-2 text--secondary">
          {{ $t('compliance.policiesDescription') }}
        </p>
      </div>
      <v-btn
        color="primary"
        @click="createDialog = true"
        :disabled="!canEdit || contents.length === 0"
      >
        <v-icon left>mdi-plus</v-icon>
        {{ $t('compliance.createPolicy') }}
      </v-btn>
    </div>

    <!-- Policies List -->
    <v-card>
      <v-data-table
        :headers="headers"
        :items="policies"
        :loading="loading"
        :items-per-page="25"
        class="elevation-1"
        @click:row="viewPolicy"
        style="cursor: pointer"
      >
        <!-- Name column -->
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center">
            <v-icon class="mr-2" color="success">mdi-shield-check</v-icon>
            <span class="font-weight-medium">{{ item.name }}</span>
          </div>
        </template>

        <!-- Content column -->
        <template v-slot:item.content="{ item }">
          <span v-if="getContentName(item.content_id)">
            {{ getContentName(item.content_id) }}
          </span>
          <span v-else class="text--secondary">{{ $t('compliance.unknownContent') }}</span>
        </template>

        <!-- Profile column -->
        <template v-slot:item.profile_id="{ item }">
          <v-chip small color="primary" outlined>
            {{ item.profile_id }}
          </v-chip>
        </template>

        <!-- Schedule column -->
        <template v-slot:item.schedule="{ item }">
          <div v-if="item.schedule_id">
            <v-icon small color="success" class="mr-1">mdi-clock</v-icon>
            <span class="text-body-2">{{ $t('compliance.scheduled') }}</span>
          </div>
          <div v-else>
            <v-icon small color="grey" class="mr-1">mdi-clock-outline</v-icon>
            <span class="text-body-2 text--secondary">{{ $t('compliance.manualOnly') }}</span>
          </div>
        </template>

        <!-- Created column -->
        <template v-slot:item.created="{ item }">
          {{ formatDate(item.created) }}
        </template>

        <!-- Actions column -->
        <template v-slot:item.actions="{ item }">
          <v-btn
            icon
            small
            @click.stop="runScan(item)"
            :title="$t('compliance.runScan')"
            color="primary"
          >
            <v-icon small>mdi-play</v-icon>
          </v-btn>
          <v-btn
            icon
            small
            @click.stop="viewPolicy(item)"
            :title="$t('compliance.viewPolicy')"
          >
            <v-icon small>mdi-eye</v-icon>
          </v-btn>
          <v-btn
            v-if="canEdit"
            icon
            small
            @click.stop="editPolicy(item)"
            :title="$t('compliance.editPolicy')"
          >
            <v-icon small>mdi-pencil</v-icon>
          </v-btn>
          <v-btn
            v-if="canEdit"
            icon
            small
            color="error"
            @click.stop="deletePolicy(item)"
            :title="$t('compliance.deletePolicy')"
          >
            <v-icon small>mdi-delete</v-icon>
          </v-btn>
        </template>

        <!-- Empty state -->
        <template v-slot:no-data>
          <div class="text-center py-8">
            <v-icon size="64" color="grey lighten-1" class="mb-4">mdi-shield-check-outline</v-icon>
            <h3 class="text-h6 mb-2">{{ $t('compliance.noPolicies') }}</h3>
            <p class="text-body-2 text--secondary mb-4">
              {{ $t('compliance.noPoliciesDescription') }}
            </p>
            <v-btn
              v-if="contents.length > 0"
              color="primary"
              @click="createDialog = true"
              :disabled="!canEdit"
            >
              <v-icon left>mdi-plus</v-icon>
              {{ $t('compliance.createFirstPolicy') }}
            </v-btn>
            <div v-else class="text-center">
              <v-btn
                color="primary"
                :to="`/project/${projectId}/compliance/contents`"
              >
                <v-icon left>mdi-file-document</v-icon>
                {{ $t('compliance.uploadContentFirst') }}
              </v-btn>
            </div>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Create/Edit Policy Dialog -->
    <v-dialog v-model="createDialog" max-width="800px" persistent>
      <v-card>
        <v-card-title>
          <span class="text-h6">
            {{ editingPolicy ? $t('compliance.editPolicy') : $t('compliance.createPolicy') }}
          </span>
        </v-card-title>

        <v-card-text>
          <v-form ref="policyForm" v-model="policyFormValid">
            <v-text-field
              v-model="policyData.name"
              :label="$t('compliance.policyName')"
              :rules="[v => !!v || $t('validation.required')]"
              required
              class="mb-4"
            />

            <v-select
              v-model="policyData.content_id"
              :items="contentOptions"
              :label="$t('compliance.selectContent')"
              :rules="[v => !!v || $t('validation.required')]"
              required
              @change="onContentChange"
              class="mb-4"
            />

            <v-select
              v-model="policyData.profile_id"
              :items="profileOptions"
              :label="$t('compliance.selectProfile')"
              :rules="[v => !!v || $t('validation.required')]"
              :disabled="!policyData.content_id"
              required
              class="mb-4"
            />

            <v-text-field
              v-model="policyData.cron_format"
              :label="$t('compliance.schedule')"
              hint="Cron format (e.g., '0 2 * * *' for daily at 2 AM)"
              persistent-hint
              placeholder="Leave empty for manual only"
            />
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn
            text
            @click="closeCreateDialog"
            :disabled="saving"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="primary"
            @click="savePolicy"
            :loading="saving"
            :disabled="!policyFormValid"
          >
            {{ editingPolicy ? $t('save') : $t('create') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Policy Details Dialog -->
    <v-dialog v-model="policyDialog" max-width="800px">
      <v-card v-if="selectedPolicy">
        <v-card-title>
          <span class="text-h6">{{ selectedPolicy.name }}</span>
          <v-spacer />
          <v-btn icon @click="policyDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <v-list dense>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.name') }}</v-list-item-title>
                    <v-list-item-subtitle>{{ selectedPolicy.name }}</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.content') }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{ getContentName(selectedPolicy.content_id) }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.profile') }}</v-list-item-title>
                    <v-list-item-subtitle>
                      <v-chip small color="primary" outlined>
                        {{ selectedPolicy.profile_id }}
                      </v-chip>
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title>{{ $t('compliance.created') }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{ formatDate(selectedPolicy.created) }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-col>
            <v-col cols="12" md="6">
              <div class="text-h6 mb-3">{{ $t('compliance.actions') }}</div>
              <v-btn
                color="primary"
                @click="runScan(selectedPolicy)"
                class="mb-2"
                block
              >
                <v-icon left>mdi-play</v-icon>
                {{ $t('compliance.runScan') }}
              </v-btn>
              <v-btn
                v-if="canEdit"
                color="secondary"
                @click="editPolicy(selectedPolicy)"
                block
              >
                <v-icon left>mdi-pencil</v-icon>
                {{ $t('compliance.editPolicy') }}
              </v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  name: 'CompliancePolicies',

  props: {
    projectId: {
      type: Number,
      required: true,
    },
    policies: {
      type: Array,
      default: () => [],
    },
    contents: {
      type: Array,
      default: () => [],
    },
  },

  data() {
    return {
      loading: false,
      createDialog: false,
      policyDialog: false,
      policyFormValid: false,
      saving: false,
      editingPolicy: false,
      selectedPolicy: null,
      policyData: {
        name: '',
        content_id: null,
        profile_id: '',
        cron_format: '',
      },
      headers: [
        { text: this.$t('compliance.name'), value: 'name', sortable: true },
        { text: this.$t('compliance.content'), value: 'content', sortable: true },
        { text: this.$t('compliance.profile'), value: 'profile_id', sortable: true },
        { text: this.$t('compliance.schedule'), value: 'schedule', sortable: false },
        { text: this.$t('compliance.created'), value: 'created', sortable: true },
        {
          text: this.$t('actions'), value: 'actions', sortable: false, width: '200px',
        },
      ],
    };
  },

  computed: {
    canEdit() {
      return this.$props.userPermissions?.includes('manageProjectResources') || false;
    },

    contentOptions() {
      return this.contents.map((content) => ({
        text: content.name,
        value: content.id,
      }));
    },

    profileOptions() {
      const content = this.contents.find((c) => c.id === this.policyData.content_id);
      if (!content || !content.profiles) {
        return [];
      }
      return content.profiles.map((profile) => ({
        text: profile.title,
        value: profile.profile_id,
      }));
    },
  },

  methods: {
    async runScan(policy) {
      try {
        await axios.post(`/api/project/${this.projectId}/compliance/policies/${policy.id}/scan`);
        this.$store.dispatch('showSuccess', this.$t('compliance.scanStarted'));
        this.$emit('refresh');
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      }
    },

    async savePolicy() {
      if (!this.$refs.policyForm.validate()) {
        return;
      }

      this.saving = true;
      try {
        const data = {
          name: this.policyData.name,
          content_id: this.policyData.content_id,
          profile_id: this.policyData.profile_id,
        };

        if (this.editingPolicy) {
          await axios.put(
            `/api/project/${this.projectId}/compliance/policies/${this.editingPolicy.id}`,
            data,
          );
        } else {
          await axios.post(`/api/project/${this.projectId}/compliance/policies`, data);
        }

        this.closeCreateDialog();
        this.$emit('refresh');
        this.$store.dispatch('showSuccess', this.$t('compliance.policySaved'));
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      } finally {
        this.saving = false;
      }
    },

    async deletePolicy(policy) {
      // eslint-disable-next-line no-alert
      if (!window.confirm(this.$t('compliance.confirmDeletePolicy', { name: policy.name }))) {
        return;
      }

      try {
        await axios.delete(`/api/project/${this.projectId}/compliance/policies/${policy.id}`);
        this.$emit('refresh');
        this.$store.dispatch('showSuccess', this.$t('compliance.policyDeleted'));
      } catch (error) {
        this.$store.dispatch('showError', getErrorMessage(error));
      }
    },

    viewPolicy(policy) {
      this.selectedPolicy = policy;
      this.policyDialog = true;
    },

    editPolicy(policy) {
      this.editingPolicy = policy;
      this.policyData = {
        name: policy.name,
        content_id: policy.content_id,
        profile_id: policy.profile_id,
        cron_format: '',
      };
      this.createDialog = true;
    },

    onContentChange() {
      this.policyData.profile_id = '';
    },

    closeCreateDialog() {
      this.createDialog = false;
      this.editingPolicy = false;
      this.resetPolicyForm();
    },

    resetPolicyForm() {
      this.policyData = {
        name: '',
        content_id: null,
        profile_id: '',
        cron_format: '',
      };
      this.$refs.policyForm?.reset();
    },

    getContentName(contentId) {
      const content = this.contents.find((c) => c.id === contentId);
      return content ? content.name : null;
    },

    formatDate(date) {
      return new Date(date).toLocaleDateString();
    },
  },
};
</script>

<style lang="scss" scoped>
.policies-container {
  // Add any custom styles here
}
</style>
