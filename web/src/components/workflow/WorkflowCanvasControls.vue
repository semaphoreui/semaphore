<template>
  <div class="WorkflowCanvasControls" @mousedown.stop @wheel.stop>
    <v-tooltip right v-for="item in items" :key="item.event">
      <template v-slot:activator="{ on, attrs }">
        <button
          type="button"
          class="WorkflowCanvasControls__btn"
          v-bind="attrs"
          v-on="on"
          @click="$emit(item.event)"
        >
          <v-icon small>{{ item.icon }}</v-icon>
        </button>
      </template>
      <span>{{ item.title }}</span>
    </v-tooltip>
    <div class="WorkflowCanvasControls__zoom">{{ zoomLabel }}</div>
  </div>
</template>

<script>
export default {
  props: {
    zoom: { type: Number, default: 1 },
    editable: { type: Boolean, default: false },
  },

  computed: {
    items() {
      const items = [
        { event: 'zoom-in', icon: 'mdi-plus', title: this.$t('workflowToolbarZoomIn') },
        { event: 'zoom-out', icon: 'mdi-minus', title: this.$t('workflowToolbarZoomOut') },
        { event: 'fit', icon: 'mdi-fit-to-page-outline', title: this.$t('workflowFitView') },
      ];
      if (this.editable) {
        items.push({ event: 'tidy', icon: 'mdi-auto-fix', title: this.$t('workflowTidyUp') });
      }
      return items;
    },
    zoomLabel() {
      return `${Math.round(this.zoom * 100)}%`;
    },
  },
};
</script>

<style lang="scss">
.WorkflowCanvasControls {
  position: absolute;
  left: 12px;
  bottom: 12px;
  z-index: 4;
  display: flex;
  flex-direction: column;
  border-radius: 8px;
  border: 1px solid var(--wf-divider);
  background: var(--wf-surface);
  box-shadow: var(--wf-shadow);
  overflow: hidden;

  &__btn {
    width: 32px;
    height: 32px;
    border: 0;
    background: transparent;
    color: var(--wf-text2);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;

    & + & {
      border-top: 1px solid var(--wf-divider);
    }

    &:hover {
      background: var(--wf-tile-bg);
    }

    .v-icon {
      color: inherit;
    }
  }

  &__zoom {
    font-size: 10px;
    text-align: center;
    color: var(--wf-text2);
    padding: 3px 0;
    border-top: 1px solid var(--wf-divider);
    user-select: none;
  }
}
</style>
