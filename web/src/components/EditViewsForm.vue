<template>
  <div v-if="views != null">
    <draggable
        v-if="views.length > 0"
        :list="views"
        handle=".handle6785"
        class="mb-5"
        @end="onDragEnd"
    >
      <div v-for="(view) in views" :key="view.id" class="d-flex mb-2">
        <v-icon class="handle6785" style="cursor: move;">mdi-menu</v-icon>
        <v-text-field
            class="ml-4 mr-1"
            hide-details
            dense
            solo
            :flat="!view.active"
            v-model="view.title"
            @focus="editView(view.id)"
            :disabled="view.disabled"
        />
        <v-btn
            class="mt-1"
            small
            icon
            @click="saveView(view.id)"
            v-if="view.active"
            :disabled="view.disabled"
        >
          <v-icon small color="green">mdi-check</v-icon>
        </v-btn>
        <v-btn
            class="mt-1"
            small
            icon
            @click="resetView(view.id)"
            v-if="view.active && view.id > 0"
            :disabled="view.disabled"
        >
          <v-icon small color="red">mdi-close</v-icon>
        </v-btn>

        <v-btn class="ml-4" icon @click="removeView(view.id)">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>
    </draggable>
    <v-alert
        v-else
        type="info"
    >{{ $t('noViews') }}</v-alert>
    <v-btn @click="addView()" color="primary">{{ $t('addView') }}</v-btn>
  </div>
</template>

<script>
import draggable from 'vuedraggable';
import axios from 'axios';

export default {
  props: {
    projectId: Number,
  },

  components: {
    draggable,
  },

  async created() {
    this.views = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/views`,
      responseType: 'json',
    })).data.map((view) => ({
      ...view,
      active: false,
      disabled: false,
    }));
    this.views.sort((v1, v2) => v1.position - v2.position);
  },

  data() {
    return {
      views: null,
    };
  },
  methods: {
    async onDragEnd() {
      const viewPositions = this.views.reduce((ret, view, index) => {
        if (view.id < 0 || view.position === index) {
          return ret;
        }
        return {
          ...ret,
          [view.id]: index,
        };
      }, {});

      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/views/positions`,
        responseType: 'json',
        data: viewPositions,
      });

      Object.keys(viewPositions).map((id) => parseInt(id, 10)).forEach((id) => {
        const view = this.views.find((v) => v.id === id);
        view.position = viewPositions[id];
      });
    },

    async saveView(viewId) {
      const i = this.views.findIndex((v) => v.id === viewId);
      if (i === -1) {
        return;
      }

      const view = this.views[i];

      if (!view.title) {
        return;
      }

      view.disabled = true;
      try {
        if (view.id < 0) {
          const newView = (await axios({
            method: 'post',
            url: `/api/project/${this.projectId}/views`,
            responseType: 'json',
            data: {
              project_id: this.projectId,
              title: view.title,
              position: i,
            },
          })).data;
          view.id = newView.id;
        } else {
          await axios({
            method: 'put',
            url: `/api/project/${this.projectId}/views/${view.id}`,
            responseType: 'json',
            data: {
              id: view.id,
              project_id: this.projectId,
              title: view.title,
              position: i,
            },
          });
        }
      } finally {
        view.disabled = false;
      }
      view.active = false;
    },

    async resetView(viewId) {
      const view = this.views.find((v) => v.id === viewId);
      if (view == null) {
        return;
      }

      view.disabled = true;
      try {
        const oldView = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/views/${view.id}`,
          responseType: 'json',
        })).data;
        view.title = oldView.title;
      } finally {
        view.disabled = false;
      }

      view.active = false;
    },

    editView(viewId) {
      const view = this.views.find((v) => v.id === viewId);
      if (view == null) {
        return;
      }
      view.active = true;
    },

    async removeView(viewId) {
      const i = this.views.findIndex((v) => v.id === viewId);
      if (i === -1) {
        return;
      }

      const view = this.views[i];

      if (view.id >= 0) {
        view.disabled = true;
        try {
          await axios({
            method: 'delete',
            url: `/api/project/${this.projectId}/views/${view.id}`,
            responseType: 'json',
          });
        } finally {
          view.disabled = false;
        }
      }

      this.views.splice(i, 1);
    },
    addView() {
      this.views.push({
        id: -Math.round(Math.random() * 10000000),
        title: '',
        active: true,
        disabled: false,
      });
    },
  },
};
</script>
