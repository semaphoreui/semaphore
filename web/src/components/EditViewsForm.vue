<template>
  <div
    v-if="views != null"
    data-testid="viewsForm-form"
    :data-state="views.length === 0 ? 'empty' : 'ready'"
  >
    <draggable
        v-if="views.length > 0"
        v-model="views"
        handle=".handle6785"
        class="mb-5"
        @end="onDragEnd"
        data-testid="viewsForm-list"
    >
      <div
        v-for="(view) in views"
        :key="view.id"
        class="mb-2"
        data-testid="viewsForm-view"
        :data-state="view.disabled ? 'loading' : 'ready'"
      >
        <div class="d-flex">

          <v-icon
            class="handle6785 align-self-center"
            style="cursor: move;"
            data-testid="viewsForm-viewHandle"
          >mdi-menu</v-icon>

          <v-text-field
            class="ml-2 mr-1 edit-views-form__title"
            hide-details
            density="compact"
            variant="solo"
            :flat="!view.active"
            v-model="view.title"
            @focus="editView(view.id)"
            :disabled="view.disabled"
            data-testid="viewsForm-title"
          />

          <v-btn
            class="mt-1"
            size="small"
            icon
            variant="text"
            @click="saveView(view.id)"
            v-if="view.active"
            :disabled="view.disabled"
            data-testid="viewsForm-viewSave"
          >
            <v-icon size="small" color="green">mdi-check</v-icon>
          </v-btn>
          <v-btn
            class="mt-1"
            size="small"
            icon
            variant="text"
            @click="resetView(view.id)"
            v-if="view.active && view.id > 0"
            :disabled="view.disabled"
            data-testid="viewsForm-viewReset"
          >
            <v-icon size="small" color="red">mdi-close</v-icon>
          </v-btn>

          <v-btn
            class="ml-4"
            icon
            variant="text"
            @click="activeViewId = (activeViewId === view.id ? null : view.id)"
            data-testid="viewsForm-viewSettings"
          >
            <v-icon>mdi-cog</v-icon>
          </v-btn>

          <v-btn
            class="ml-1"
            icon
            variant="text"
            @click="removeView(view.id)"
            data-testid="viewsForm-viewDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </div>

        <!-- overflow: visible — a Vuetify 3 card clips its content; the corner above ⚙
             sticks out of it -->
        <v-card
          v-if="view.id === activeViewId"
          style="background: var(--highlighted-card-bg-color); overflow: visible;"
          class="mb-6 pt-3 mt-5"
          data-testid="viewsForm-viewSettingsCard"
        >

          <div style="
            position: absolute;
            background: var(--highlighted-card-bg-color);
            width: 28px;
            height: 28px;
            transform: rotate(45deg);
            right: 45px;
            top: -14px;
            border-radius: 0;
          "></div>

          <v-card-text>
            <v-select
              :items="viewTypes"
              item-value="slug"
              item-title="title"
              v-model="view.type"
              label="Type"
              variant="outlined"
              density="compact"
              @update:model-value="saveView(view.id)"
              data-testid="viewsForm-type"
            />

            <div
              class="d-flex"
              style="align-items: center; column-gap: 12px;"
            >

              <v-select
                :items="sortableColumns"
                item-value="name"
                item-title="title"
                v-model="view.sort_column"
                label="Sort by"
                variant="outlined"
                density="compact"
                hide-details
                @update:model-value="saveView(view.id)"
                data-testid="viewsForm-sortColumn"
              />

              <v-checkbox
                hide-details
                label="Reverse"
                v-model="view.sort_reverse"
                class="mt-0 pt-0"
                @update:model-value="saveView(view.id)"
                data-testid="viewsForm-sortReverse"
              />

            </div>

            <v-checkbox
              hide-details
              label="Hidden"
              v-model="view.hidden"
              @update:model-value="saveView(view.id)"
              data-testid="viewsForm-hidden"
            />
          </v-card-text>

        </v-card>

      </div>

    </draggable>
    <v-alert
        v-else
        type="info"
        data-testid="viewsForm-empty"
    >{{ $t('noViews') }}</v-alert>
    <v-btn
      @click="addView()"
      color="primary"
      data-testid="viewsForm-add"
    >{{ $t('addView') }}</v-btn>
  </div>
</template>

<style lang="scss">
// The view title: a dense solo field of Vuetify 2 is 38 px high
// (`.v-text-field--solo.v-input--dense` control), not the 40 px of a compact field; its text line
// (20 px, padding 4/2 in a 26 px input 6 px from the top) has its middle 20 px from the top.
.edit-views-form__title .v-field {
  --v-input-control-height: 38px;
  --v-input-padding-top: 8px;
  --v-field-padding-bottom: 6px;
}
</style>

<script>
import { VueDraggable as draggable } from 'vue-draggable-plus';
import axios from 'axios';
// import ArgsPicker from '@/components/ArgsPicker.vue';

export default {
  props: {
    projectId: Number,
  },

  components: {
    // ArgsPicker,
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
      type: view.type ? view.type : '',
    }));
    this.views.sort((v1, v2) => v1.position - v2.position);
  },

  data() {
    return {
      views: null,
      activeViewId: null,
      viewTypes: [{
        slug: '',
        title: 'Custom',
      }, {
        slug: 'all',
        title: 'All',
      }],
      sortableColumns: [{
        name: 'name',
        title: 'Name',
      }],
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
              type: view.type,
              sort_column: view.sort_column,
              sort_reverse: view.sort_reverse,
              hidden: view.hidden,
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
              type: view.type,
              sort_column: view.sort_column,
              sort_reverse: view.sort_reverse,
              hidden: view.hidden,
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
        type: '',
        active: true,
        disabled: false,
      });
    },
  },
};
</script>
