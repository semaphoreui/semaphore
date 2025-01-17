<template>
    <pre v-if="state" style="white-space: pre-wrap;
            background: gray;
            color: white;
            border-radius: 10px;
            overflow: auto;
            font-size: 14px;
            max-height: 400px;
            margin-top: 5px;"
         class="pa-2"
    >{{ state.state }}</pre>
    <div v-else-if="error" class="text-center">{{ error.message }}</div>
</template>

<script>
import axios from 'axios';

export default {
  props: {
    projectId: Number,
    inventoryId: Number,
    stateId: Number,
  },

  data() {
    return {
      state: null,
      error: null,
    };
  },

  async created() {
    try {
      this.state = (await axios.get(`/api/project/${this.projectId}/inventory/${this.inventoryId}/terraform/states/${this.stateId}`)).data;
    } catch (e) {
      if (e.response.status === 404) {
        this.error = {
          message: 'No state available.',
        };
      } else {
        this.error = e;
      }
    }
  },
};
</script>
