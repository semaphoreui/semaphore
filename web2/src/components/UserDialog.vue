<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    max-width="400"
    persistent
    :transition="false"
  >
    <v-card>
      <v-card-title class="headline">{{ isNew ? 'New' : 'Edit' }} User</v-card-title>

      <v-card-text>
        <UserForm :user-id="userId" ref="form" />
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
          {{ isNew ? 'Create' : 'Save' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

import UserForm from '@/components/UserForm.vue';
import EventBus from '@/event-bus';

export default {
  components: { UserForm },
  props: {
    userId: [Number, String],
    value: Boolean,
  },

  data() {
    return {
      dialog: false,
    };
  },

  computed: {
    isNew() {
      return this.userId === 'new';
    },
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
      if (await this.$refs.form) {
        await this.$refs.form.reset();
      }
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
    async save() {
      const item = await this.$refs.form.save();
      if (!item) {
        return null;
      }

      this.$emit('saved', {
        item,
        action: this.isNew ? 'new' : 'edit',
      });

      EventBus.$emit('i-user', {
        action: this.isNew ? 'new' : 'edit',
        item,
      });

      this.dialog = false;
      return item;
    },
  },
};
</script>
