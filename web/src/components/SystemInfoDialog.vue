<template>
  <v-dialog
    v-model="dialog"
    max-width="700"
    scrollable
  >
    <v-card>
      <v-card-title class="headline">
        {{ $t('systemInfo') }}
        <v-spacer />
        <v-btn icon @click="dialog = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-card-text v-if="loading" class="text-center py-8">
        <v-progress-circular indeterminate color="primary" />
      </v-card-text>

      <v-card-text v-else-if="error" class="py-4">
        <v-alert type="error" dense outlined>
          {{ error }}
        </v-alert>
      </v-card-text>

      <v-card-text v-else-if="info" class="py-2">

        <!-- System -->
        <v-subheader class="px-0">{{ $t('system') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('version') }}
              </td>
              <td>{{ info.system.version }}</td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('goVersion') }}</td>
              <td>
                {{ info.system.go_version }}
                ({{ info.system.go_os }}/{{ info.system.go_arch }})
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('gitClient') }}</td>
              <td>{{ info.system.git_client }}</td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('tmpPath') }}</td>
              <td>{{ info.system.tmp_path }}</td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('homeDirMode') }}</td>
              <td>{{ info.system.home_dir_mode }}</td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('scheduleTimezone') }}</td>
              <td>{{ info.system.schedule_timezone }}</td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Ansible -->
        <v-subheader class="px-0 mt-2">Ansible</v-subheader>
        <v-simple-table dense v-if="info.system.ansible">
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('version') }}
              </td>
              <td>
                <pre class="ansible-version">{{ info.system.ansible.trim() }}</pre>
              </td>
            </tr>
          </tbody>
        </v-simple-table>
        <div v-else class="px-0 text--secondary text-body-2">
          {{ $t('ansibleNotFound') }}
        </div>

        <!-- Database -->
        <v-subheader class="px-0 mt-2">{{ $t('database') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('dbDialect') }}
              </td>
              <td>{{ info.database.dialect }}</td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Authentication -->
        <v-subheader class="px-0 mt-2">{{ $t('authentication') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('passwordLogin') }}
              </td>
              <td>
                <v-icon small :color="info.auth.password_login_enabled ? 'success' : 'grey'">
                  {{ info.auth.password_login_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">TOTP</td>
              <td>
                <v-icon small :color="info.auth.totp_enabled ? 'success' : 'grey'">
                  {{ info.auth.totp_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('emailOtp') }}</td>
              <td>
                <v-icon small :color="info.auth.email_otp_enabled ? 'success' : 'grey'">
                  {{ info.auth.email_otp_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">LDAP</td>
              <td>
                <v-icon small :color="info.auth.ldap_enabled ? 'success' : 'grey'">
                  {{ info.auth.ldap_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('oidcProviders') }}</td>
              <td>
                <span v-if="info.auth.oidc_providers && info.auth.oidc_providers.length > 0">
                  {{ info.auth.oidc_providers.join(', ') }}
                </span>
                <span v-else class="text--secondary">{{ $t('none') }}</span>
              </td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Notifications -->
        <v-subheader class="px-0 mt-2">{{ $t('notifications') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr v-for="(enabled, name) in info.notifications" :key="name">
              <td class="font-weight-medium" style="width: 200px; text-transform: capitalize">
                {{ formatNotificationName(name) }}
              </td>
              <td>
                <v-icon small :color="enabled ? 'success' : 'grey'">
                  {{ enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Cluster / HA -->
        <v-subheader class="px-0 mt-2">{{ $t('cluster') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('highAvailability') }}
              </td>
              <td>
                <v-icon small :color="info.cluster.ha_enabled ? 'success' : 'grey'">
                  {{ info.cluster.ha_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr v-if="info.cluster.ha_enabled">
              <td class="font-weight-medium">{{ $t('nodeId') }}</td>
              <td>{{ info.cluster.node_id }}</td>
            </tr>
            <tr v-if="info.cluster.ha_enabled && info.cluster.node_count != null">
              <td class="font-weight-medium">{{ $t('nodeCount') }}</td>
              <td>{{ info.cluster.node_count }}</td>
            </tr>
            <tr v-if="info.cluster.ha_enabled && info.cluster.ui_count != null">
              <td class="font-weight-medium">{{ $t('uiInstances') }}</td>
              <td>{{ info.cluster.ui_count }}</td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Runners -->
        <v-subheader class="px-0 mt-2">{{ $t('runners') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('useRemoteRunner') }}
              </td>
              <td>
                <v-icon small :color="info.runners.use_remote_runner ? 'success' : 'grey'">
                  {{ info.runners.use_remote_runner ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
            <tr v-if="info.runners.total != null">
              <td class="font-weight-medium">{{ $t('globalRunners') }}</td>
              <td>{{ info.runners.active }} / {{ info.runners.total }} {{ $t('active2') }}</td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Task Settings -->
        <v-subheader class="px-0 mt-2">{{ $t('taskSettings') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('maxParallelTasks') }}
              </td>
              <td>{{ info.task_settings.max_parallel_tasks || $t('unlimited') }}</td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('maxTaskDuration') }}</td>
              <td>
                {{ info.task_settings.max_task_duration_sec
                  ? info.task_settings.max_task_duration_sec + 's'
                  : $t('unlimited') }}
              </td>
            </tr>
            <tr>
              <td class="font-weight-medium">{{ $t('maxTasksPerTemplate') }}</td>
              <td>{{ info.task_settings.max_tasks_per_template || $t('unlimited') }}</td>
            </tr>
          </tbody>
        </v-simple-table>

        <!-- Features -->
        <v-subheader class="px-0 mt-2">{{ $t('featureFlags') }}</v-subheader>
        <v-simple-table dense>
          <tbody>
            <tr>
              <td class="font-weight-medium" style="width: 200px">
                {{ $t('nonAdminCanCreateProject') }}
              </td>
              <td>
                <v-icon
                  small
                  :color="info.features.non_admin_can_create_project
                    ? 'success' : 'grey'"
                >
                  {{ info.features.non_admin_can_create_project
                    ? 'mdi-check-circle' : 'mdi-close-circle' }}
                </v-icon>
              </td>
            </tr>
          </tbody>
        </v-simple-table>

      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn color="blue darken-1" text @click="dialog = false">
          {{ $t('close') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.ansible-version {
  font-size: 12px;
  white-space: pre-wrap;
  margin: 0;
  font-family: monospace;
}
</style>

<script>
import axios from 'axios';

export default {
  props: {
    value: Boolean,
  },

  data() {
    return {
      dialog: false,
      info: null,
      loading: false,
      error: null,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
      if (val && !this.info) {
        await this.loadInfo();
      }
    },

    value(val) {
      this.dialog = val;
    },
  },

  methods: {
    formatNotificationName(name) {
      return name.replace(/_/g, ' ');
    },

    async loadInfo() {
      this.loading = true;
      this.error = null;
      try {
        this.info = (await axios({
          method: 'get',
          url: '/api/admin/info',
          responseType: 'json',
        })).data;
      } catch (err) {
        this.error = err.response?.data?.message || err.message || 'Failed to load system info';
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
