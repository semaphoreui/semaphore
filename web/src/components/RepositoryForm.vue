<template>
  <v-form
      ref="form"
      v-model="formValid"
      v-if="item != null && keys != null"
      data-testid="repository-form"
      :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
        :model-value="!!formError"
        color="error"
        class="pb-2"
        data-testid="repository-error"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        :label="$t('name')"
        :rules="[v => !!v || $t('name_required')]"
        required
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="repository-name"
    ></v-text-field>

    <v-text-field
        v-model.trim="item.git_url"
        :label="$t('urlOrPath')"
        :rules="[
          v => !!v || $t('repository_required'),
          v => getTypeOfUrl(v) != null || $t('incorrectUrl'),
        ]"
        required
        :disabled="formSaving"
        :hide-details="true"
        variant="outlined"
        density="compact"
        data-testid="repository-gitUrl"
    ></v-text-field>

    <div class="mt-1 mb-5" data-testid="repository-urlTypes">
      <span class="text-caption">git:</span>
      <v-chip
        v-for="x in ['ssh', 'http', 'https', 'file', 'git']"
        size="x-small"
        class="ml-1"
        :color="type ===x ? 'primary' : ''"
        @click="setType(x)"
        style="font-weight: bold;"
        :key="x"
        :data-testid="`repository-urlType-${x}`"
      >{{ x }}</v-chip>
      <span class="text-caption ml-3">local:</span>
      <v-chip
        size="x-small"
        class="ml-1"
        :color="type ==='local' ? 'primary' : ''"
        @click="setType('local')"
        style="font-weight: bold;"
        data-testid="repository-urlType-local"
      >
        {{ $t('absPath') }}
      </v-chip>
    </div>

    <v-text-field
      v-model.trim="item.git_branch"
      :label="$t('branch')"
      :rules="[v => (!!v || type === 'local') || $t('branch_required')]"
      required
      :disabled="formSaving || type === 'local'"
      variant="outlined"
      density="compact"
      data-testid="repository-gitBranch"
    ></v-text-field>

    <v-checkbox
      v-model="item.monorepo"
      :label="$t('monorepo')"
      :hint="$t('monorepoHelp')"
      persistent-hint
      :disabled="formSaving || type === 'local'"
      class="mt-0 mb-4"
      data-testid="repository-monorepo"
    ></v-checkbox>

    <v-autocomplete
        v-model="item.ssh_key_id"
        :label="$t('accessKey')"
        :items="keys"
        item-value="id"
        item-title="name"
        :rules="[v => !!v || $t('key_required')]"
        required
        :disabled="formSaving"
        variant="outlined"
        density="compact"
        data-testid="repository-sshKeyId"
    >
      <template v-slot:append>
        <v-tooltip location="left" content-class="opacity1 bg-black">
          <template v-slot:activator="{ props }">
            <v-icon
              v-bind="props"
              data-testid="repository-sshKeyHelp"
            >
              mdi-help-circle
            </v-icon>
          </template>
          <div class="py-4" data-testid="repository-sshKeyHelpText">
            <p>{{ $t('credentialsToAccessToTheGitRepositoryItShouldBe') }}</p>
            <!-- pl-6: the list indent of Vuetify 2 (.v-application ul); Vuetify 3 has none,
                 without it the bullets fall out of the tooltip box -->
            <ul class="pl-6">
              <li><code>{{ $t('ssh') }}</code> {{ $t('ifYouUseGitOrSshUrl') }}</li>
              <li><code>{{ $t('none') }}</code> {{ $t('ifYouUseHttpsOrFileUrl') }}</li>
            </ul>
          </div>
        </v-tooltip>
      </template>
    </v-autocomplete>
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';
import { getRepositoryUrlType } from '@/lib/repositoryUrl';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      helpDialog: null,
      helpKey: null,

      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'static-yaml',
        name: 'Static YAML',
      }, {
        id: 'file',
        name: 'File',
      }],
    };
  },
  async created() {
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },
  computed: {
    type() {
      return getRepositoryUrlType(this.item.git_url);
    },
  },

  methods: {
    getTypeOfUrl(url) {
      return getRepositoryUrlType(url);
    },

    setType(type) {
      let url;

      const m = this.item.git_url.match(/^\w+:\/\/(.*)$/);
      if (m != null) {
        url = m[1];
      } else {
        url = this.item.git_url;
      }

      if (type === 'local') {
        const isWinDrive = /^[a-zA-Z]:/.test(url);
        const isUnc = /^\\\\/.test(url);
        if (!url.startsWith('/') && !isWinDrive && !isUnc) {
          url = `/${url}`;
        }
      } else {
        url = `${type}://${url}`;
      }

      this.item.git_url = url;
    },

    showHelpDialog(key) {
      this.helpKey = key;
      this.helpDialog = true;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/repositories`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/repositories/${this.itemId}`;
    },
  },
};
</script>
