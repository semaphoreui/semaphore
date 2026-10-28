<template>
  <v-dialog v-model="dialog" max-width="700" scrollable>
    <v-card
      data-testid="systemInfoDialog"
      :data-state="loading ? 'loading' : error ? 'error' : info ? 'ready' : 'empty'"
    >
      <v-card-title class="headline" data-testid="systemInfoDialog-title">
        {{ $t('systemInfo') }}
        <v-spacer />
        <v-btn icon @click="dialog = false" data-testid="systemInfoDialog-close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-card-text v-if="loading" class="text-center py-8" data-testid="systemInfo-loading">
        <v-progress-circular indeterminate color="primary" />
      </v-card-text>

      <v-card-text v-else-if="error" class="py-4">
        <v-alert type="error" dense outlined data-testid="systemInfo-error">
          {{ error }}
        </v-alert>
      </v-card-text>

      <v-card-text v-else-if="info" class="py-2 mb-4" data-testid="systemInfo-content">
        <!-- System -->
        <v-subheader class="px-0">{{ $t('system') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-system"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('version') }}
                  </td>
                  <td data-testid="systemInfo-version">{{ info.system.version }}</td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('goVersion') }}</td>
                  <td data-testid="systemInfo-goVersion">
                    {{ info.system.go_version }}
                    ({{ info.system.go_os }}/{{ info.system.go_arch }})
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('gitClient') }}</td>
                  <td data-testid="systemInfo-gitClient">{{ info.system.git_client }}</td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('tmpPath') }}</td>
                  <td data-testid="systemInfo-tmpPath">{{ info.system.tmp_path }}</td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('homeDirMode') }}</td>
                  <td data-testid="systemInfo-homeDirMode">{{ info.system.home_dir_mode }}</td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('scheduleTimezone') }}</td>
                  <td
                    data-testid="systemInfo-scheduleTimezone"
                  >{{ info.system.schedule_timezone }}</td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Ansible -->
        <v-subheader class="px-0 mt-2">Ansible</v-subheader>
        <v-card
          data-testid="systemInfo-section-ansible"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="py-2">
            <pre
              v-if="info.system.ansible"
              class="ansible-version"
              data-testid="systemInfo-ansibleVersion"
            >
              {{ info.system.ansible.trim() }}
            </pre>
            <div
              v-else
              class="px-0 text--secondary text-body-2"
              data-testid="systemInfo-ansibleNotFound"
            >
              {{ $t('ansibleNotFound') }}
            </div>
          </v-card-text>
        </v-card>

        <!-- Database -->
        <v-subheader class="px-0 mt-2">{{ $t('database') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-database"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('dbDialect') }}
                  </td>
                  <td data-testid="systemInfo-dbDialect">{{ info.database.dialect }}</td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Authentication -->
        <v-subheader class="px-0 mt-2">{{ $t('authentication') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-auth"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('passwordLogin') }}
                  </td>
                  <td data-testid="systemInfo-passwordLogin">
                    <v-icon small :color="info.auth.password_login_enabled ? 'success' : 'grey'">
                      {{
                        info.auth.password_login_enabled ? 'mdi-check-circle' : 'mdi-close-circle'
                      }}
                    </v-icon>
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">TOTP</td>
                  <td data-testid="systemInfo-totp">
                    <v-icon small :color="info.auth.totp_enabled ? 'success' : 'grey'">
                      {{ info.auth.totp_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                    </v-icon>
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('emailOtp') }}</td>
                  <td data-testid="systemInfo-emailOtp">
                    <v-icon small :color="info.auth.email_otp_enabled ? 'success' : 'grey'">
                      {{
                        info.auth.email_otp_enabled ? 'mdi-check-circle' : 'mdi-close-circle'
                      }}
                    </v-icon>
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">LDAP</td>
                  <td data-testid="systemInfo-ldap">
                    <v-icon small :color="info.auth.ldap_enabled ? 'success' : 'grey'">
                      {{ info.auth.ldap_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                    </v-icon>
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('oidcProviders') }}</td>
                  <td data-testid="systemInfo-oidcProviders">
                    <span v-if="info.auth.oidc_providers && info.auth.oidc_providers.length > 0">
                      {{ info.auth.oidc_providers.join(', ') }}
                    </span>
                    <span v-else class="text--secondary">{{ $t('none') }}</span>
                  </td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Notifications -->
        <v-subheader class="px-0 mt-2">{{ $t('notifications') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-notifications"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr
                  v-for="(enabled, name) in info.notifications"
                  :key="name"
                  :data-testid="`systemInfo-notification-${name}`"
                >
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
          </v-card-text>
        </v-card>

        <v-subheader class="px-0 mt-2">{{ $t('cluster') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-cluster"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('highAvailability') }}
                  </td>
                  <td data-testid="systemInfo-haEnabled">
                    <v-icon small :color="info.cluster.ha_enabled ? 'success' : 'grey'">
                      {{ info.cluster.ha_enabled ? 'mdi-check-circle' : 'mdi-close-circle' }}
                    </v-icon>
                  </td>
                </tr>
                <tr v-if="info.cluster.ha_enabled">
                  <td class="font-weight-medium">{{ $t('nodeId') }}</td>
                  <td data-testid="systemInfo-nodeId">{{ info.cluster.node_id }}</td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Runners -->
        <v-subheader class="px-0 mt-2">{{ $t('runners') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-runners"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('useRemoteRunner') }}
                  </td>
                  <td data-testid="systemInfo-useRemoteRunner">
                    <v-icon small :color="info.runners.use_remote_runner ? 'success' : 'grey'">
                      {{ info.runners.use_remote_runner ? 'mdi-check-circle' : 'mdi-close-circle' }}
                    </v-icon>
                  </td>
                </tr>
                <tr v-if="info.runners.total != null">
                  <td class="font-weight-medium">{{ $t('globalRunners') }}</td>
                  <td
                    data-testid="systemInfo-globalRunners"
                  >{{ info.runners.active }} / {{ info.runners.total }} {{ $t('active2') }}</td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Task Settings -->
        <v-subheader class="px-0 mt-2">{{ $t('taskSettings') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-taskSettings"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('maxParallelTasks') }}
                  </td>
                  <td
                    data-testid="systemInfo-maxParallelTasks"
                  >{{ info.task_settings.max_parallel_tasks || $t('unlimited') }}</td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('maxTaskDuration') }}</td>
                  <td data-testid="systemInfo-maxTaskDuration">
                    {{
                      info.task_settings.max_task_duration_sec
                        ? info.task_settings.max_task_duration_sec + 's'
                        : $t('unlimited')
                    }}
                  </td>
                </tr>
                <tr>
                  <td class="font-weight-medium">{{ $t('maxTasksPerTemplate') }}</td>
                  <td
                    data-testid="systemInfo-maxTasksPerTemplate"
                  >{{ info.task_settings.max_tasks_per_template || $t('unlimited') }}</td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>

        <!-- Features -->
        <v-subheader class="px-0 mt-2">{{ $t('featureFlags') }}</v-subheader>
        <v-card
          data-testid="systemInfo-section-features"
          style="background: var(--highlighted-card-bg-color)"
        >
          <v-card-text class="px-0 py-2">
            <v-simple-table dense style="background: transparent">
              <tbody>
                <tr>
                  <td class="font-weight-medium" style="width: 200px">
                    {{ $t('nonAdminCanCreateProject') }}
                  </td>
                  <td data-testid="systemInfo-nonAdminCanCreateProject">
                    <v-icon
                      small
                      :color="info.features.non_admin_can_create_project ? 'success' : 'grey'"
                    >
                      {{
                        info.features.non_admin_can_create_project
                          ? 'mdi-check-circle'
                          : 'mdi-close-circle'
                      }}
                    </v-icon>
                  </td>
                </tr>
              </tbody>
            </v-simple-table>
          </v-card-text>
        </v-card>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.ansible-version {
  font-size: 12px;
  margin: 0;
  font-family: monospace;
  overflow-x: auto;
}
</style>

<script>
import axios from 'axios';

export default {
  props: {
    value: Boolean,
  },

  emits: ['input'],

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
        this.info = (
          await axios({
            method: 'get',
            url: '/api/admin/info',
            responseType: 'json',
          })
        ).data;
      } catch (err) {
        this.error = err.response?.data?.message || err.message || 'Failed to load system info';
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
