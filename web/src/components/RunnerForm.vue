<template>
  <v-form ref="form" lazy-validation v-model="formValid" v-if="item != null">
    <v-alert :value="formError" color="error" class="pb-2">{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <div style="position: relative">
      <v-combobox
        v-model="item.tags"
        :label="$t('Tags')"
        :items="tagSuggestions || []"
        :rules="
          projectId
            ? [(v) => item.is_default || (Array.isArray(v) && v.length > 0) || $t('tag_required')]
            : []
        "
        :required="!!projectId"
        :disabled="formSaving || !isTagsAvailable"
        :loading="tagSuggestions == null"
        multiple
        chips
        deletable-chips
        small-chips
        hide-selected
        outlined
        hide-details
      />

      <v-chip
        v-if="!isTagsAvailable"
        color="hsl(348deg, 86%, 61%)"
        text-color="white"
        small
        label
        style="position: absolute; top: -10px; right: 15px"
        @click="upgradeToPro()"
      >
        Upgrade to PRO
      </v-chip>
    </div>

    <v-checkbox label="Is default" v-model="item.is_default" />

    <v-text-field
      v-model="item.webhook"
      :label="$t('Webhook')"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-text-field
      type="number"
      v-model.number="item.max_parallel_tasks"
      :label="$t('maxNumberOfParallelTasksOptional')"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-checkbox
      style="position: absolute; left: 24px; bottom: 15px"
      class="mt-0"
      v-model="item.active"
      :label="$t('enabled')"
      :disabled="formSaving"
      hide-details
    ></v-checkbox>
  </v-form>
</template>
<script>
import axios from 'axios';
import ItemFormBase from '@/components/ItemFormBase';
import EventBus from '@/event-bus';

export default {
  props: {
    isAdmin: Boolean,
    projectId: Number,
    isTagsAvailable: {
      type: Boolean,
      default: true,
    },
  },

  mixins: [ItemFormBase],

  data() {
    return {
      tagSuggestions: null,
    };
  },

  async created() {
    try {
      const url = this.projectId
        ? `/api/project/${this.projectId}/runner_tags`
        : '/api/runner_tags';
      const { data } = await axios.get(url);
      // The endpoint returns [{tag, number_of_runners}]; v-combobox just needs strings.
      this.tagSuggestions = Array.isArray(data) ? data.map((t) => t.tag) : [];
    } catch (err) {
      this.tagSuggestions = [];
    }
  },

  methods: {
    upgradeToPro() {
      EventBus.$emit('i-subscription', {});
    },

    getItemsUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/runners`;
      }

      return '/api/runners';
    },

    beforeSave() {
      if (!this.item.max_parallel_tasks) {
        this.item.max_parallel_tasks = 0;
      }

      // v-combobox emits the typed token only after blur — coerce to trimmed,
      // de-duped strings so the API never sees blank or duplicate tags.
      const seen = new Set();
      this.item.tags = (this.item.tags || [])
        .map((t) => (typeof t === 'string' ? t.trim() : ''))
        .filter((t) => {
          if (!t || seen.has(t)) {
            return false;
          }
          seen.add(t);
          return true;
        });
    },

    getSingleItemUrl() {
      if (this.projectId) {
        return `/api/project/${this.projectId}/runners/${this.itemId}`;
      }
      return `/api/runners/${this.itemId}`;
    },
  },
};
</script>
