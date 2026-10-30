<template>
  <v-form
      ref="form"
      v-model="formValid"
      v-if="item != null && keys != null"
      class="RepositoryForm"
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
      <span class="text-body-small">git:</span>
      <v-chip
        v-for="x in ['ssh', 'http', 'https', 'file', 'git']"
        size="x-small"
        class="ml-1 RepositoryForm__urlType"
        :color="type ===x ? 'primary' : ''"
        @click="setType(x)"
        style="font-weight: bold;"
        :key="x"
        :data-testid="`repository-urlType-${x}`"
      >{{ x }}</v-chip>
      <span class="text-body-small ml-3">local:</span>
      <v-chip
        size="x-small"
        class="ml-1 RepositoryForm__urlType"
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
        class="RepositoryForm__key"
        data-testid="repository-sshKeyId"
    >
      <template v-slot:append>
        <!-- viewport-margin 4: on a phone the tooltip keeps the width of Vuetify 2, which pushed
             it only off the left edge (12 px), not off both -->
        <v-tooltip location="left" content-class="opacity1 bg-black" viewport-margin="4">
          <template v-slot:activator="{ props }">
            <v-icon
              v-bind="props"
              data-testid="repository-sshKeyHelp"
            >
              mdi-help-circle
            </v-icon>
          </template>
          <div class="py-4" data-testid="repository-sshKeyHelpText">
            <!-- mb-4: the paragraph margin of Vuetify 2 (.v-application p) -->
            <p class="mb-4">{{ $t('credentialsToAccessToTheGitRepositoryItShouldBe') }}</p>
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
<style lang="scss">
// The URL type chips: an x-small chip of Vuetify 2 is 16 px high (`.v-chip.v-size--x-small`),
// Vuetify 4 makes it 20 px, which also made the row and the dialog 0.3 px higher.
.RepositoryForm__urlType.v-chip {
  --v-chip-height: 16px;
}

// The help icon after the key field (`#append-outer` of Vuetify 2): 9 px from the field
// (`.v-input__append-outer` margin-left), not 16 px, so the field and its menu keep their v2
// width; an icon of the theme (`.theme--light.v-icon`: .54, white in the dark theme) in every
// state — not the medium emphasis twice of Vuetify 4, not red on a field error, not dimmed again
// while saving (App.vue already fades the whole disabled field to .5).
.RepositoryForm__key > .v-input__append {
  margin-inline-start: 9px;
  opacity: 1;

  > .v-icon {
    opacity: 1;
  }
}

.RepositoryForm__key.v-theme--light > .v-input__append > .v-icon {
  color: rgba(0, 0, 0, .54);
}

.RepositoryForm__key.v-theme--dark > .v-input__append > .v-icon {
  color: #fff;
}

// A disabled field (saving, a local path): the hint of Monorepo keeps its colour, the checkbox
// and its label are in the disabled text colour (`.v-input--is-disabled` of Vuetify 2) — under
// the .5 of App.vue, not dimmed to .38 once more by Vuetify 4.
.RepositoryForm .v-input--disabled > .v-input__details {
  opacity: 1;
}

.RepositoryForm .v-selection-control--disabled {
  opacity: 1;

  .v-label,
  .v-selection-control__input > .v-icon {
    color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    opacity: 1;
  }
}
</style>
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
