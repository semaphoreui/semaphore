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
        <v-list-item-title class="WorkflowVarAssign__empty text-wrap">
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

    // The subheader of a dense list of Vuetify 2 (`.v-list--dense .v-subheader`): 12 px, 8 px in.
    > .v-list-subheader {
      font-size: .75rem;
      line-height: 18px;
      padding: 0 8px;
      background: none;
    }
  }

  // The disabled «no variables» item: the disabled text colour of Vuetify 2
  // (`.theme--light.v-list-item--disabled`), not the dimmed item of Vuetify 4; its caption text
  // (`text-caption`) 16 px a line, as every title of a dense list (`.v-list--dense
  // .v-list-item__title`: the class did not set the line height with `!important`).
  // (App.vue sets 1rem on the titles of a dense list; the caption class won with `!important`).
  &__list > .v-list-item--disabled {
    opacity: 1;
    color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));

    .v-list-item-title.WorkflowVarAssign__empty {
      font-size: .75rem;
      line-height: 16px;
      letter-spacing: .0333333333em;
    }
  }

  // An item of a dense list of Vuetify 2 with a title and a subtitle: 8 px above and below the
  // text (`.v-list--dense .v-list-item__content`), title 16 px a line and 2 px under it.
  &__item.v-list-item {
    padding-top: 8px;
    padding-bottom: 8px;

    .v-list-item-title {
      line-height: 16px;
      margin-bottom: 2px;
    }
  }

  // The subtitle of a dense list of Vuetify 2 (`.v-list--dense .v-list-item__subtitle`) overrode
  // the 11 px of this class: 13 px, medium weight, 16 px a line, in the secondary text colour.
  &__name {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: .8125rem;
    font-weight: 500;
    line-height: 16px;
    opacity: 1;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
  }
}
</style>
