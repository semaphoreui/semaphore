import { APP_FIELDS, UNKNOWN_APP_FIELDS } from '@/lib/constants';

export default {

  computed: {
    fields() {
      return APP_FIELDS[this.app] || UNKNOWN_APP_FIELDS;
    },
  },

  methods: {

    fieldLabel(f) {
      return this.$t((this.fields[f] || { label: f }).label);
    },

    needField(f) {
      return this.fields[f] != null;
    },

    isFieldRequired(f) {
      return this.fields[f] != null && !this.fields[f].optional;
    },
  },
};
