<template>
  <div class="pt-3">
    <v-alert dense outlined type="success">
      {{ $t('projectWithNameRestored', { projectName: project.name }) }}
    </v-alert>

    <v-alert dense outlined type="error" class="mb-0">
      <b>{{ $t('emptyKeysRestored', { emptyKeys }) }}</b>
      {{ $t('pleaseUpdateAccessKeys') }}
    </v-alert>

    <template v-if="rows.length > 0">
      <div class="text-subtitle-1 mt-4">{{ $t('unassignedRolesTitle') }}</div>

      <v-alert v-if="loadError" dense outlined type="error" class="mt-2 mb-0">
        {{ loadError }}
      </v-alert>

      <div v-for="row in rows" :key="row.slug" class="mt-3">
        <div>
          <b>{{ row.slug }}</b>
          <span v-if="row.name" class="ml-1">{{ row.name }}</span>
        </div>
        <div class="text--secondary">{{ row.templates.map((tpl) => tpl.name).join(', ') }}</div>

        <v-select
          v-model="choices[row.slug]"
          :items="itemsFor(row)"
          item-value="slug"
          item-text="name"
          :label="$t('assignToRole')"
          outlined
          dense
          hide-details
          class="mt-2"
          :disabled="assigning"
        >
          <template v-slot:item="{ item: role }">
            <v-list-item-content>
              <v-list-item-title>{{ role.name }}</v-list-item-title>
              <v-list-item-subtitle v-if="role.slug && role.slug !== CREATE">
                {{ role.slug }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </template>
        </v-select>

        <v-text-field
          v-if="choices[row.slug] === CREATE"
          v-model="newSlugs[row.slug]"
          :label="$t('slug')"
          :rules="[v => isValidSlug(v) || $t('invalidSlugFormat')]"
          outlined
          dense
          class="mt-2"
          :disabled="assigning"
        />

        <v-alert
          v-if="errors[row.slug]"
          dense
          outlined
          type="error"
          class="mt-2 mb-0"
          style="white-space: pre-line"
        >{{ errors[row.slug] }}</v-alert>
      </div>

      <v-btn
        color="primary"
        class="mt-3"
        :disabled="!hasChoices"
        :loading="assigning"
        @click="assign()"
      >{{ $t('assign') }}</v-btn>
    </template>
  </div>
</template>
<script>
import axios from 'axios';
import groupRolesForSelect, { isValidSlug } from '@/lib/roles';
import { getErrorMessage } from '@/lib/error';

const CREATE = '__create__';

function stateFor(project) {
  const unassigned = project.unassigned_roles || [];
  return {
    roles: [],
    loadError: null,
    choices: Object.fromEntries(unassigned.map((row) => [row.slug, null])),
    newSlugs: Object.fromEntries(unassigned.map((row) => [row.slug, ''])),
    errors: {},
    granted: {},
    done: {},
    assigning: false,
  };
}

/** What a restore left for the user: a choice of role per unassigned role from the backup file. */
export default {
  props: {
    project: Object,
    emptyKeys: Number,
  },

  data() {
    return { CREATE, ...stateFor(this.project) };
  },

  computed: {
    rows() {
      return (this.project.unassigned_roles || []).filter((row) => !this.done[row.slug]);
    },

    hasChoices() {
      return this.rows.some((row) => this.choiceFor(row) !== null);
    },
  },

  watch: {
    project() {
      Object.assign(this.$data, stateFor(this.project));
      this.loadRoles();
    },
  },

  created() {
    this.loadRoles();
  },

  methods: {
    isValidSlug,

    async loadRoles() {
      try {
        this.roles = (await axios.get(`/api/project/${this.project.id}/roles/all`)).data;
      } catch (err) {
        this.loadError = getErrorMessage(err);
        return;
      }
      this.rows.forEach((row) => {
        if (this.roles.some((role) => role.slug === row.slug)) {
          this.choices[row.slug] = row.slug;
        }
      });
    },

    itemsFor(row) {
      const items = [{ slug: null, name: this.$t('dontAssign') }];
      if (row.reason === 'slug_taken') {
        items.push({ slug: CREATE, name: this.$t('createAsNewRole') });
      }
      return items.concat(groupRolesForSelect(this.roles, (key) => this.$i18n.t(key)));
    },

    /** The slug the row's templates go to, or null when there is nothing valid to do yet. */
    choiceFor(row) {
      const choice = this.choices[row.slug];
      if (choice === CREATE) {
        return isValidSlug(this.newSlugs[row.slug]) ? this.newSlugs[row.slug] : null;
      }
      return choice || null;
    },

    async assign() {
      this.assigning = true;
      const rows = this.rows.filter((row) => this.choiceFor(row) !== null);
      for (let i = 0; i < rows.length; i += 1) {
        // eslint-disable-next-line no-await-in-loop
        await this.assignRow(rows[i]);
      }
      this.assigning = false;
    },

    async assignRow(row) {
      const base = `/api/project/${this.project.id}`;
      const roleSlug = this.choiceFor(row);
      if (this.choices[row.slug] === CREATE) {
        try {
          await axios.post(`${base}/roles`, {
            name: row.name,
            slug: roleSlug,
            permissions: row.permissions,
          });
        } catch (err) {
          this.$set(this.errors, row.slug, getErrorMessage(err));
          return;
        }
        this.roles.push({ slug: roleSlug, name: row.name, project_id: this.project.id });
        this.choices[row.slug] = roleSlug;
      }

      const failed = [];
      for (let i = 0; i < row.templates.length; i += 1) {
        const tpl = row.templates[i];
        const key = `${row.slug}/${tpl.id}`;
        if (!this.granted[key]) {
          try {
            // eslint-disable-next-line no-await-in-loop
            await axios.post(`${base}/templates/${tpl.id}/perms`, {
              role_slug: roleSlug,
              permissions: tpl.permissions,
            });
            this.$set(this.granted, key, true);
          } catch (err) {
            failed.push(`${tpl.name}: ${getErrorMessage(err)}`);
          }
        }
      }
      if (failed.length > 0) {
        this.$set(this.errors, row.slug, failed.join('\n'));
        return;
      }
      this.$delete(this.errors, row.slug);
      this.$set(this.done, row.slug, true);
    },
  },
};
</script>
