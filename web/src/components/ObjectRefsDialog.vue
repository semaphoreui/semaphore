<template>
  <v-dialog
    v-model="dialog"
    max-width="400"
    :close-on-back="false"
  >
    <v-card :data-testid="testId">
      <v-card-title
        class="text-headline-medium font-weight-medium ObjectRefsDialog__title"
        data-testid="objectRefsDialog-title"
      >{{ $t('cantDeleteThe', {objectTitle: objectTitle}) }}
      </v-card-title>

      <v-card-text class="pb-0">
        <ObjectRefsView
          data-testid="objectRefs-view"
          :object-title="objectTitle"
          :object-refs="objectRefs"
          :project-id="projectId"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="blue-darken-1"
          variant="text"
          @click="dialog = false"
          data-testid="objectRefsDialog-close"
        >{{ $t('close2') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<style lang="scss">
// A long title wraps, breaking a word if needed (`.v-card__title` of Vuetify 2: flex-wrap,
// word-break: break-all), instead of the one ellipsised line of Vuetify 4.
.ObjectRefsDialog__title {
  white-space: normal;
  word-break: break-all;
}
</style>
<script>
import ObjectRefsView from '@/components/ObjectRefsView.vue';

export default {
  components: {
    ObjectRefsView,
  },
  props: {
    modelValue: Boolean,
    objectRefs: Object,
    projectId: Number,
    objectTitle: String,
    testId: {
      type: String,
      default: 'objectRefsDialog',
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      dialog: false,
    };
  },
  watch: {
    async dialog(val) {
      this.$emit('update:modelValue', val);
    },

    async modelValue(val) {
      this.dialog = val;
    },
  },
};
</script>
