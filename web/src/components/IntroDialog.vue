<template>
  <EditDialog
    v-model="dialog"
    :max-width="1000"
    hide-buttons
    :expandable="true"
    no-body-paddings
    @close="onClose()"
    test-id="introDialog"
    title="New release v2.17 🎉"
  >
    <template v-slot:form="{}">
      <v-carousel>
        <v-carousel-item>

        </v-carousel-item>
        <v-carousel-item> </v-carousel-item>
      </v-carousel>
    </template>
  </EditDialog>
</template>
<script>
import EditDialog from '@/components/EditDialog.vue';
import ProjectMixin from '@/components/ProjectMixin';

export default {
  components: { EditDialog },

  mixins: [ProjectMixin],

  props: {
    value: Boolean,
    projectId: Number,
    itemId: Number,
    systemInfo: Object,
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.item = null;
      this.template = null;
      this.dialog = val;
      await this.loadData();
    },
  },

  data() {
    return {
      dialog: null,
      model: 0,
    };
  },

  methods: {
    close() {
      this.dialog = false;
      this.onClose();
    },

    async loadData() {
      //
    },

    onClose() {
      this.$emit('close');
    },
  },
};
</script>
