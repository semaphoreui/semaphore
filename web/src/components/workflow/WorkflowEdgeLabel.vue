<template>
  <div
    class="WorkflowEdgeLabel"
    :class="[
      `WorkflowEdgeLabel--${label.condition}`,
      label.state ? `WorkflowEdgeLabel--${label.state}` : '',
    ]"
    :style="{ left: `${label.x}px`, top: `${label.y}px` }"
    @mousedown.stop
    @touchstart.stop
  >
    <v-menu v-if="editable" offset-y nudge-top="-4" transition="fade-transition">
      <template v-slot:activator="{ on, attrs }">
        <span
          class="WorkflowEdgeLabel__pill WorkflowEdgeLabel__pill--clickable"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon x-small :color="current.color">{{ current.icon }}</v-icon>
          <span class="WorkflowEdgeLabel__text">{{ current.text }}</span>
          <button
            type="button"
            class="WorkflowEdgeLabel__del"
            :title="$t('workflowEdgeDelete')"
            @mousedown.stop
            @click.stop="$emit('remove')"
          >
            <v-icon x-small>mdi-close</v-icon>
          </button>
        </span>
      </template>
      <v-list dense>
        <v-list-item
          v-for="opt in options"
          :key="opt.value"
          :input-value="opt.value === label.condition"
          @click="$emit('change', opt.value)"
        >
          <v-list-item-icon class="mr-2">
            <v-icon small :color="opt.color">{{ opt.icon }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ opt.text }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>

    <span v-else class="WorkflowEdgeLabel__pill">
      <v-icon x-small :color="current.color">{{ current.icon }}</v-icon>
      <span class="WorkflowEdgeLabel__text">{{ current.text }}</span>
    </span>
  </div>
</template>

<script>
// Condition pill drawn on the midpoint of an edge. Positioned in canvas
// coordinates inside the Drawflow canvas, so it pans and zooms with it.
export default {
  props: {
    // { key, source, dest, condition, x, y, state }
    label: { type: Object, required: true },
    editable: { type: Boolean, default: false },
  },

  computed: {
    options() {
      return [
        {
          value: 'on_success', icon: 'mdi-check', color: 'success', text: this.$t('workflowConditionOnSuccess'),
        },
        {
          value: 'on_failure', icon: 'mdi-close', color: 'error', text: this.$t('workflowConditionOnFailure'),
        },
        {
          value: 'always', icon: 'mdi-arrow-right', color: 'grey', text: this.$t('workflowConditionAlways'),
        },
      ];
    },
    current() {
      return this.options.find((o) => o.value === this.label.condition) || this.options[0];
    },
  },
};
</script>

<style lang="scss">
.WorkflowEdgeLabel {
  position: absolute;
  transform: translate(-50%, -50%);
  pointer-events: auto;
  z-index: 1;

  &__pill {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    height: 20px;
    padding: 0 8px 0 6px;
    border-radius: 10px;
    font-size: 10.5px;
    font-weight: 500;
    letter-spacing: 0.02em;
    white-space: nowrap;
    background: var(--wf-surface);
    border: 1px solid var(--wf-divider);
    box-shadow: var(--wf-shadow);
    color: var(--wf-text);
    user-select: none;

    &--clickable {
      cursor: pointer;

      &:hover {
        border-color: var(--wf-divider-strong);
        box-shadow: var(--wf-shadow-hover);
      }
    }
  }

  &__del {
    display: none;
    align-items: center;
    margin-left: 2px;
    padding: 0 0 0 5px;
    border: 0;
    border-left: 1px solid var(--wf-divider);
    background: transparent;
    color: var(--wf-text2);
    cursor: pointer;
    line-height: 1;
  }

  &__pill:hover &__del {
    display: inline-flex;
  }

  &--dim {
    opacity: 0.55;
  }
}
</style>
