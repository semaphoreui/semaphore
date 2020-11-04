<template>
  <div>
    <v-container class="pa-0" v-if="item != null && output != null">
      <v-row no-gutters>
        <v-col>
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Author</v-list-item-title>
                <v-list-item-subtitle>{{ item.user_name }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Status</v-list-item-title>
                <v-list-item-subtitle>{{ item.status }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col>
          <v-list two-line subheader class="pa-0">
            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Created</v-list-item-title>
                <v-list-item-subtitle>{{ item.created }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Started</v-list-item-title>
                <v-list-item-subtitle>{{ item.start }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>

            <v-list-item class="pa-0">
              <v-list-item-content>
                <v-list-item-title>Ended</v-list-item-title>
                <v-list-item-subtitle>{{ item.end }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>
    </v-container>

    <div class="text-view" style="height: 400px;" v-text="output">
    </div>
  </div>
</template>
<style lang="scss">
  .text-view {
    overflow: auto;
    border: 1px solid gray;
    border-radius: 4px;
    font-family: monospace;
  }
</style>
<script>
import axios from 'axios';

export default {
  props: {
    itemId: Number,
    projectId: Number,
  },
  data() {
    return {
      item: null,
      output: null,
    };
  },
  async created() {
    await this.loadData();
  },
  methods: {
    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}`,
        responseType: 'json',
      })).data;
      this.output = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.itemId}/output`,
        responseType: 'json',
      })).data.map((line) => line.output).join('\n');
    },
  },
};
</script>
