<template>
  <div v-if="showField" class="mb-4">
    <div class="d-flex align-center mb-3">
      <h3>
        {{ $t('additionalRepositories') }}
        <v-chip class="ml-2" small color="error">New</v-chip>
      </h3>
      <v-btn
        small
        color="primary"
        class="ml-3"
        @click="addRepo"
        :disabled="formSaving"
      >
        <v-icon small left>mdi-plus</v-icon>
        {{ $t('add') }}
      </v-btn>
    </div>

    <v-row>
      <v-col
        v-for="(addRepo, index) in repositories"
        :key="index"
        cols="12"
        md="6"
      >
        <fieldset
          class="mb-3"
          style="padding: 10px;
                 border-width: 1px;
                 border-color: rgba(133, 133, 133, 0.4);
                 background-color: rgba(133, 133, 133, 0.1);
                 border-radius: 8px;
                 font-size: 12px;"
        >
          <legend
            style="
              padding: 0 3px;
              width: 100%;
              display: flex;
              justify-content: space-between;
              align-items: center;
            "
          >
            <span>{{ $t('repository') }} #{{ index + 1 }}</span>
            <v-btn
              icon
              small
              color="error"
              @click="removeRepo(index)"
              :disabled="formSaving"
            >
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </legend>

          <v-autocomplete
            v-model="addRepo.repository_id"
            :label="$t('repository') + ' *'"
            :items="gitRepositories"
            item-value="id"
            item-text="name"
            :rules="[v => !!v || $t('repository_required')]"
            outlined
            dense
            hide-details
            :disabled="formSaving"
            class="mb-4"
            @change="onRepositoryChange(index)"
          ></v-autocomplete>

          <v-text-field
            v-model="addRepo.path"
            :label="$t('path') + ' *'"
            :rules="[
              v => !!v || $t('path_required'),
              v => /^[a-zA-Z0-9_/-]+$/.test(v) || $t('pathInvalidCharacters'),
              v => !isDuplicatePath(v, index) || $t('pathAlreadyUsed'),
            ]"
            outlined
            dense
            hide-details
            :disabled="formSaving"
            :placeholder="$t('exampleMyRepo')"
            @input="slugifyPath(index)"
            prefix="repos/"
          ></v-text-field>

          <div class="mb-3 text-right">
            <a
              v-if="!addRepo.git_branch && !addRepo.showBranch"
              @click="setShowBranch(index, true)"
            >{{ $t('setBranch') }}</a>
          </div>

          <div v-if="addRepo.git_branch || addRepo.showBranch">
            <v-autocomplete
              v-if="branches[index] && branches[index].length > 0"
              clearable
              v-model="addRepo.git_branch"
              :label="$t('branch')"
              :items="branches[index]"
              outlined
              dense
              :disabled="formSaving"
              :placeholder="$t('optional')"
            ></v-autocomplete>

            <v-text-field
              v-else
              clearable
              v-model="addRepo.git_branch"
              :label="$t('branch')"
              outlined
              dense
              :disabled="formSaving"
              :placeholder="$t('optional')"
            ></v-text-field>

            <div class="text-right" style="margin-top: -12px; margin-bottom: 12px;">
              <a
                v-if="addRepo.showBranch"
                @click="
                  setShowBranch(index, false);
                  addRepo.git_branch = null;
                "
              >{{ $t('useDefaultBranch') }}</a>
            </div>
          </div>
        </fieldset>
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Array,
      default: () => [],
    },
    gitRepositories: {
      type: Array,
      default: () => [],
    },
    branches: {
      type: Object,
      default: () => ({}),
    },
    formSaving: {
      type: Boolean,
      default: false,
    },
    showField: {
      type: Boolean,
      default: true,
    },
  },

  computed: {
    repositories: {
      get() {
        return this.value || [];
      },
      set(val) {
        this.$emit('input', val);
      },
    },
  },

  methods: {
    addRepo() {
      const repos = [...this.repositories];
      repos.push({
        id: 0,
        repository_id: null,
        path: '',
        git_branch: null,
        position: repos.length,
        showBranch: false,
      });
      this.repositories = repos;
    },

    removeRepo(index) {
      const repos = [...this.repositories];
      repos.splice(index, 1);
      // Update positions
      for (let i = 0; i < repos.length; i += 1) {
        repos[i].position = i;
      }
      this.repositories = repos;
    },

    setShowBranch(index, value) {
      const repos = [...this.repositories];
      repos[index].showBranch = value;
      this.repositories = repos;
    },

    slugifyPath(index) {
      const repos = [...this.repositories];
      const repo = repos[index];
      if (repo.path) {
        let slugified = repo.path
          .replace(/^\/+|\/+$/g, '') // Remove leading and trailing slashes FIRST
          .toLowerCase()
          .replace(/[^a-z0-9_/-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '');
        // Remove slashes again at the end to catch any edge cases
        slugified = slugified.replace(/^\/+|\/+$/g, '');
        repos[index].path = slugified;
        this.repositories = repos;
      }
    },

    onRepositoryChange(index) {
      this.$emit('load-branches', index);
    },

    isDuplicatePath(path, currentIndex) {
      if (!path) return false;
      return this.repositories.some((repo, idx) => (
        idx !== currentIndex && repo.path === path
      ));
    },
  },
};
</script>
