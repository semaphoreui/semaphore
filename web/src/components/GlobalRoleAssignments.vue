<template>
  <div>
    <v-alert :value="!!error" color="error" class="pb-2">{{ error }}</v-alert>

    <v-alert
      v-if="loaded && availableRoles.length === 0 && assignments.length === 0"
      text
      dense
      type="info"
    >
      {{ $t('noGlobalRolesAvailable') }}
    </v-alert>

    <v-row align="center" v-if="availableRoles.length > 0">
      <v-col cols="12" sm="8">
        <v-select
          v-model="newRoleId"
          :items="availableRoles"
          item-value="id"
          item-text="name"
          :label="$t('globalRole')"
          outlined
          dense
          hide-details
          :disabled="loading"
          data-testid="global-role-select"
        />
      </v-col>
      <v-col cols="12" sm="4">
        <v-btn
          block
          color="primary"
          :disabled="!newRoleId || loading"
          :loading="loading"
          @click="assign()"
        >
          {{ $t('assignGlobalRole') }}
        </v-btn>
      </v-col>
    </v-row>

    <v-list v-if="assignments.length > 0" class="px-0">
      <v-list-item v-for="assignment in assignments" :key="assignment.id" class="px-0">
        <v-list-item-content>
          <v-list-item-title>{{ assignment.role_name }}</v-list-item-title>
          <TemplatePermissionsChips
            class="pt-1"
            scope="global"
            :permissions="assignment.global_permissions || 0"
          />
        </v-list-item-content>
        <v-list-item-action>
          <v-btn
            icon
            :disabled="loading"
            :aria-label="$t('removeGlobalRole', { role: assignment.role_name })"
            @click="remove(assignment)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </v-list-item-action>
      </v-list-item>
    </v-list>
    <v-alert v-else-if="availableRoles.length > 0" text dense type="info" class="mt-4">
      {{ $t('noGlobalRolesAssigned') }}
    </v-alert>

    <v-subheader class="px-0 mt-4">{{ $t('effectiveGlobalPermissions') }}</v-subheader>
    <v-list dense class="px-0">
      <v-list-item
        v-for="decision in effective.decisions || []"
        :key="decision.permission"
        class="px-0"
      >
        <v-list-item-icon class="mr-3">
          <v-icon :color="decision.allowed ? 'success' : 'grey'">
            {{ decision.allowed ? 'mdi-check-circle' : 'mdi-minus-circle-outline' }}
          </v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>{{ permissionTitle(decision.permission) }}</v-list-item-title>
          <v-list-item-subtitle>{{ provenance(decision) }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </div>
</template>

<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import { ROLE_PERMISSIONS } from '@/lib/constants';
import TemplatePermissionsChips from '@/components/TemplatePermissionsChips.vue';

/**
 * The global roles a user holds and what they add up to. Needs the signed-in user to hold
 * "Can manage global roles": every request below is behind that permission.
 */
export default {
  components: { TemplatePermissionsChips },

  props: {
    userId: [Number, String],
  },

  data() {
    return {
      roles: [],
      catalog: [],
      assignments: [],
      effective: { permissions: 0, decisions: [] },
      newRoleId: null,
      loading: false,
      loaded: false,
      error: null,
    };
  },

  computed: {
    /** Global roles the user does not hold yet. */
    availableRoles() {
      const held = new Set(this.assignments.map((assignment) => assignment.role_id));
      return this.roles.filter((role) => !held.has(role.id));
    },
  },

  watch: {
    async userId() {
      await this.load();
    },
  },

  async created() {
    await this.load();
  },

  methods: {
    async get(url) {
      return (await axios({ method: 'get', url, responseType: 'json' })).data || [];
    },

    async load() {
      this.error = null;
      this.loading = true;
      try {
        [this.roles, this.catalog] = await Promise.all([
          this.get('/api/global-roles'),
          this.get('/api/global-roles/permissions'),
        ]);
        await this.loadUser();
      } catch (err) {
        this.error = getErrorMessage(err);
      } finally {
        this.loading = false;
        this.loaded = true;
      }
    },

    async loadUser() {
      [this.assignments, this.effective] = await Promise.all([
        this.get(`/api/users/${this.userId}/global-roles`),
        this.get(`/api/users/${this.userId}/global-permissions`),
      ]);
    },

    async assign() {
      if (!this.newRoleId) {
        return;
      }
      await this.change(() => axios({
        method: 'post',
        url: `/api/users/${this.userId}/global-roles`,
        responseType: 'json',
        data: { role_id: this.newRoleId },
      }));
      this.newRoleId = null;
    },

    async remove(assignment) {
      await this.change(() => axios({
        method: 'delete',
        url: `/api/users/${this.userId}/global-roles/${assignment.id}`,
        responseType: 'json',
      }));
    },

    async change(request) {
      this.error = null;
      this.loading = true;
      try {
        await request();
        await this.loadUser();
      } catch (err) {
        this.error = getErrorMessage(err);
      } finally {
        this.loading = false;
      }
    },

    permissionTitle(permissionId) {
      const definition = this.catalog.find((entry) => entry.id === permissionId);
      const known = definition && ROLE_PERMISSIONS.global.find(
        (entry) => entry.permission === definition.permission,
      );
      if (known) {
        return this.$t(known.label);
      }
      return definition ? definition.description : permissionId;
    },

    provenance(decision) {
      if (!decision.allowed) {
        return this.$t('notGranted');
      }
      if (decision.role_id) {
        return this.$t('grantedByRole', { role: decision.role_name });
      }
      return this.$t('grantedByAdministrator');
    },
  },
};
</script>
