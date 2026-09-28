<template>
  <div class="WorkflowEdgeProperties">
    <div class="WorkflowEdgeProperties__header">
      <span class="WorkflowEdgeProperties__tile">
        <v-icon>mdi-ray-start-arrow</v-icon>
      </span>
      <div class="WorkflowEdgeProperties__heading">
        <div class="text-subtitle-2">{{ $t('workflowEdgeCondition') }}</div>
        <div class="text-caption text--secondary">
          #{{ edge.source_node_id }} → #{{ edge.destination_node_id }}
        </div>
      </div>
      <v-btn icon small :title="$t('close')" @click="$emit('close')">
        <v-icon small>mdi-close</v-icon>
      </v-btn>
    </div>

    <v-divider />

    <div class="pa-4">
      <v-select
        v-model="edge.condition"
        :items="conditionOptions"
        item-value="value"
        item-text="text"
        :disabled="!canManage"
        outlined
        dense
        hide-details="auto"
        @change="$emit('input', edge)"
      >
        <template v-slot:item="{ item }">
          <v-icon small left :color="item.color">{{ item.icon }}</v-icon>
          {{ item.text }}
        </template>
        <template v-slot:selection="{ item }">
          <v-icon small left :color="item.color">{{ item.icon }}</v-icon>
          {{ item.text }}
        </template>
      </v-select>
    </div>

    <v-divider />

    <div class="pa-4">
      <v-btn
        outlined
        small
        block
        color="error"
        :disabled="!canManage"
        @click="$emit('delete')"
      >
        <v-icon left small>mdi-delete-outline</v-icon>
        {{ $t('workflowEdgeDelete') }}
      </v-btn>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    // { source_node_id, destination_node_id, condition }
    value: { type: Object, required: true },
    canManage: { type: Boolean, default: false },
  },

  data() {
    return { edge: this.value };
  },

  watch: {
    value(v) {
      this.edge = v;
    },
  },

  computed: {
    conditionOptions() {
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
  },
};
</script>

<style lang="scss">
.WorkflowEdgeProperties {
  &__header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 12px 12px 16px;
  }

  &__tile {
    flex: 0 0 36px;
    width: 36px;
    height: 36px;
    border-radius: 9px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(127, 127, 127, 0.12);
  }

  &__heading {
    flex: 1 1 auto;
    min-width: 0;
  }
}
</style>
