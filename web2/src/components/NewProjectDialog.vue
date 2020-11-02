<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    max-width="400"
    persistent
    :transition="false"
  >
    <v-card>
      <v-card-title class="headline">New Project</v-card-title>

      <v-card-text>
        <ProjectEditForm project-id="new" ref="form" />
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn
          color="blue darken-1"
          text
          @click="dialog = false"
        >
          Cancel
        </v-btn>

        <v-btn
          color="blue darken-1"
          text
          @click="save()"
        >
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

import ProjectEditForm from '@/components/ProjectEditForm.vue';
import EventBus from '@/event-bus';

export default {
  components: { ProjectEditForm },
  props: {
    projectId: [Number, String],
    value: Boolean,
  },

  data() {
    return {
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
      if (await this.$refs.form) {
        await this.$refs.form.reset();
      }
    },

    async value(val) {
      if (!val) {
        this.dialog = val;
        return;
      }
      this.dialog = val;
    },
  },

  methods: {
    async save() {
      const item = await this.$refs.form.save();
      if (!item) {
        return;
      }
      EventBus.$emit('i-project', {
        action: 'new',
        item,
      });
      this.dialog = false;
    },
  },
};
</script>
