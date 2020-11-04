<template>
  <v-container>
    <v-row>
      <v-col>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Author</v-list-item-title>
              <v-list-item-subtitle>{{ item.user_name }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>

          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Status</v-list-item-title>
              <v-list-item-subtitle>{{ item.status }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Created</v-list-item-title>
              <v-list-item-subtitle>{{ item.created }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>

          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Started</v-list-item-title>
              <v-list-item-subtitle>{{ item.start }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>

          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Ended</v-list-item-title>
              <v-list-item-subtitle>{{ item.end }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>
  </v-container>
</template>
<script>
import axios from 'axios';

export default {
  props: {
    taskId: Number,
    projectId: Number,
  },
  async created() {
    await this.loadData();
  },
  methods: {
    async loadData() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.taskId}`,
        responseType: 'json',
      })).data;
      this.output = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks/${this.taskId}/output`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
