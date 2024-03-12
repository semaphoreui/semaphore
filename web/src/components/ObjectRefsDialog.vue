<template>
  <v-dialog
    v-model="dialog"
    max-width="400"
  >
    <v-card>
      <v-card-title class="headline">{{ $t('cantDeleteThe', {objectTitle: objectTitle}) }}
      </v-card-title>

      <v-card-text class="pb-0">
        <ObjectRefsView
          :object-title="objectTitle"
          :object-refs="objectRefs"
          :project-id="projectId"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="blue darken-1"
          text
          @click="dialog = false"
        >{{ $t('close2') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import ObjectRefsView from '@/components/ObjectRefsView.vue';

export default {
  components: {
    ObjectRefsView,
  },
  props: {
    value: Boolean,
    objectRefs: Object,
    projectId: Number,
    objectTitle: String,
  },
  data() {
    return {
      dialog: false,
    };
  },
  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.dialog = val;
    },
  },
};
</script>
