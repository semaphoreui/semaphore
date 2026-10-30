<template>
  <v-menu location="bottom end" eager>
    <template v-slot:activator="{ props }">
      <v-btn
        icon
        variant="text"
        size="small"
        :title="$t('workflowNodeSurveyInputsAssign')"
        class="WorkflowVarAssign__btn"
        :data-testid="`${testId}-open`"
        v-bind="props"
      >
        <v-icon size="small">mdi-link-variant-plus</v-icon>
      </v-btn>
    </template>
    <v-list density="compact" class="WorkflowVarAssign__list" :data-testid="`${testId}-list`">
      <v-list-subheader
        :data-testid="`${testId}-title`"
      >{{ $t('workflowNodeSurveyInputsAssignTitle') }}</v-list-subheader>
      <v-list-item v-if="vars.length === 0" disabled :data-testid="`${testId}-empty`">
        <v-list-item-title class="text-body-small text-wrap">
          {{ $t('workflowNodeSurveyInputsNoWorkflowVars') }}
        </v-list-item-title>
      </v-list-item>
      <v-list-item
        v-for="v in vars"
        :key="v.name"
        class="WorkflowVarAssign__item"
        :data-testid="`${testId}-item-${v.name}`"
        @click="$emit('select', v.name)"
      >
        <v-list-item-title>{{ v.title || v.name }}</v-list-item-title>
        <v-list-item-subtitle class="WorkflowVarAssign__name">
          {{ v.name }} · {{ v.type || 'string' }}
        </v-list-item-subtitle>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script>
// The "assign a workflow variable" button next to a survey field of a workflow
// task node: a menu of the workflow's variables; picking one emits `select`
// with its name. The parent decides what the binding means.
export default {
  props: {
    // Non-secret survey variables of the workflow: { name, title, type }.
    vars: { type: Array, default: () => [] },
    // Builds data-testid of the button, the menu and its items only, no behaviour.
    testId: { type: String, default: 'workflowVarAssign' },
  },
  emits: ['select'],
};
</script>

<style lang="scss">
.WorkflowVarAssign {
  &__list {
    min-width: 220px;
    max-width: 320px;
  }

  &__name {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 11px;
  }
}
</style>
