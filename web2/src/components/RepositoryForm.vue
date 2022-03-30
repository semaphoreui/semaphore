<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="item != null && keys != null"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        label="Name"
        :rules="[v => !!v || 'Name is required']"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.git_url"
        label="URL or path"
        :rules="[
          v => !!v || 'Repository is required',
          v => getTypeOfUrl(v) != null || 'Incorrect URL',
        ]"
        required
        :disabled="formSaving"
        :hide-details="true"
    ></v-text-field>

    <div class="mt-1 mb-4">
      <span class="caption">git:</span>
      <v-chip
        x-small
        class="ml-1"
        :color="type ==='file' ? 'primary' : ''"
        @click="setType('file')"
        style="font-weight: bold;"
      >
        file
      </v-chip>
      <v-chip
        x-small
        class="ml-1"
        :color="type ==='git' ? 'primary' : ''"
        @click="setType('git')"
        style="font-weight: bold;"
      >
        git
      </v-chip>
      <v-chip
        x-small
        class="ml-1"
        :color="type ==='ssh' ? 'primary' : ''"
        @click="setType('ssh')"
        style="font-weight: bold;"
      >
        ssh
      </v-chip>
      <span class="caption ml-3">local:</span>
      <v-chip
        x-small
        class="ml-1"
        :color="type ==='local' ? 'primary' : ''"
        @click="setType('local')"
        style="font-weight: bold;"
      >
        abs. path
      </v-chip>
    </div>

    <v-text-field
      v-model="item.git_branch"
      label="Branch"
      :rules="[v => !!v || 'Branch is required']"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-select
        v-model="item.ssh_key_id"
        label="Access Key"
        append-outer-icon="mdi-help-circle"
        :items="keys"
        item-value="id"
        item-text="name"
        :rules="[v => !!v || 'Key is required']"
        required
        :disabled="formSaving"
        @click:append-outer="showHelpDialog('key')"
    ></v-select>

    <v-dialog
        v-model="helpDialog"
        hide-overlay
        width="300"
    >
      <v-alert
          border="top"
          colored-border
          type="info"
          elevation="2"
          class="mb-0"
      >
        <div v-if="helpKey === 'url'">
          <p>
            Address of the repository with your Ansible playbooks. It can be:
          </p>
          <ul>
            <li>Git URL <code>git://</code></li>
            <li>SSH URL <code>ssh://</code></li>
            <li>HTTPS URL <code>https://</code></li>
            <li>file URL <code>file://</code></li>
          </ul>
        </div>
        <div v-else-if="helpKey === 'key'">
          <p>Credentials to access to the Git repository. It should be:</p>
          <ul>
            <li><code>SSH</code> if you use Git or SSH URL.</li>
            <li><code>None</code> if you use HTTPS or file URL.</li>
          </ul>
        </div>
      </v-alert>
    </v-dialog>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      helpDialog: null,
      helpKey: null,

      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'file',
        name: 'File',
      }],
    };
  },
  async created() {
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },
  computed: {
    type() {
      return this.getTypeOfUrl(this.item.git_url);
    },
  },

  methods: {
    getTypeOfUrl(url) {
      if (url == null || url === '') {
        return null;
      }

      if (url.startsWith('/')) {
        return 'local';
      }

      const m = url.match(/^(\w+):\/\//);

      if (m == null) {
        return 'ssh';
      }

      if (!['git', 'file', 'ssh'].includes(m[1])) {
        return null;
      }

      return m[1];
    },

    setType(type) {
      let url;

      const m = this.item.git_url.match(/^\w+:\/\/(.*)$/);
      if (m != null) {
        url = m[1];
      } else {
        url = this.item.git_url;
      }

      if (type === 'local') {
        url = url.startsWith('/') ? url : `/${url}`;
      } else {
        url = `${type}://${url}`;
      }

      this.item.git_url = url;
    },

    showHelpDialog(key) {
      this.helpKey = key;
      this.helpDialog = true;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/repositories`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/repositories/${this.itemId}`;
    },
  },
};
</script>
