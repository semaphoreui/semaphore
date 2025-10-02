<template>
  <div>
    <v-card>
      <v-card-title class="d-flex justify-space-between align-center">
        <div class="d-flex align-center">
          <v-icon class="mr-2">mdi-code-json</v-icon>
          JSON Template Content
        </div>
        <div>
          <v-btn
            color="primary"
            small
            @click="copyToClipboard"
            class="mr-2"
          >
            <v-icon left>mdi-content-copy</v-icon>
            Copy JSON
          </v-btn>
          <v-btn
            color="secondary"
            small
            @click="downloadJson"
          >
            <v-icon left>mdi-download</v-icon>
            Download JSON
          </v-btn>
        </div>
      </v-card-title>
      <v-card-text>
        <v-textarea
          :value="formattedJson"
          readonly
          auto-grow
          rows="20"
          class="json-content"
          outlined
          dense
          hide-details
        ></v-textarea>
      </v-card-text>
    </v-card>

    <v-card class="mt-4">
      <v-card-title>
        <v-icon class="mr-2">mdi-information</v-icon>
        Template Information
      </v-card-title>
      <v-card-text>
        <v-row>
          <v-col cols="12" md="6">
            <v-text-field
              :value="template.name"
              label="Template Name"
              readonly
              outlined
              dense
            />
          </v-col>
          <v-col cols="12" md="6">
            <v-text-field
              :value="template.app"
              label="Template Type"
              readonly
              outlined
              dense
            />
          </v-col>
          <v-col cols="12">
            <v-textarea
              :value="template.description"
              label="Description"
              readonly
              outlined
              dense
              rows="2"
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </div>
</template>

<script>
export default {
  name: 'TemplateJsonView',
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  computed: {
    formattedJson() {
      try {
        // Try to parse and format the JSON from the playbook field
        const jsonData = JSON.parse(this.template.playbook);
        return JSON.stringify(jsonData, null, 2);
      } catch (error) {
        // If parsing fails, return the raw content
        return this.template.playbook || '';
      }
    },
  },
  methods: {
    async copyToClipboard() {
      try {
        await navigator.clipboard.writeText(this.formattedJson);
        this.$toast?.success('JSON copied to clipboard!');
      } catch (error) {
        console.error('Failed to copy to clipboard:', error);
        this.$toast?.error('Failed to copy to clipboard');
      }
    },

    downloadJson() {
      try {
        const blob = new Blob([this.formattedJson], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${this.template.name.replace(/[^a-z0-9]/gi, '_').toLowerCase()}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        this.$toast?.success('JSON file downloaded!');
      } catch (error) {
        console.error('Failed to download JSON:', error);
        this.$toast?.error('Failed to download JSON file');
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.json-content >>> .v-text-field__slot textarea {
  font-family: 'Courier New', monospace !important;
  font-size: 14px !important;
  line-height: 1.4 !important;
}
</style>
