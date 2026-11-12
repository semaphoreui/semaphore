<template>
  <v-textarea
    v-if="templateParams.allow_override_prompt"
    :model-value="prompt"
    @update:model-value="setPrompt"
    :label="$t('agentPrompt') + (hasTemplatePrompt ? '' : ' *')"
    :rules="[(v) => !!(v && v.trim()) || hasTemplatePrompt || $t('agent_prompt_required')]"
    :hint="$t('agentTaskPromptHint', { expression: AGENT_PROMPT_EXPRESSION })"
    persistent-hint
    rows="4"
    auto-grow
    variant="outlined"
    density="compact"
    :data-testid="testId"
  ></v-textarea>
</template>
<script>
// The prompt of an AI agent task, for a template which lets a task replace its prompt
// (`allow_override_prompt`); the server refuses a task prompt otherwise. The field starts with
// the template's prompt. A prompt left equal to it, or emptied, is not sent: the task — or the
// schedule, integration or workflow node keeping these params — then runs with the template's
// prompt as it is at the time of the run. A prompt sent from here goes to the agent as it is
// written: the ${{ … }} expressions are replaced only in the prompt kept in the template.
import { AGENT_PROMPT_EXPRESSION } from '@/lib/constants';

export default {
  props: {
    modelValue: Object,
    templateParams: Object,
    testId: {
      type: String,
      default: 'taskParams-agentPrompt',
    },
  },

  emits: ['update:modelValue'],

  data() {
    return {
      // What the field shows, apart from the params: an emptied field stays empty while the
      // params fall back to the template's prompt.
      prompt: this.modelValue?.prompt ?? this.templateParams?.prompt ?? '',
      AGENT_PROMPT_EXPRESSION,
    };
  },

  computed: {
    hasTemplatePrompt() {
      return !!(this.templateParams?.prompt || '').trim();
    },
  },

  watch: {
    'modelValue.prompt': function syncPrompt(value) {
      if (value != null && value !== this.prompt) {
        this.prompt = value;
      }
    },
  },

  methods: {
    setPrompt(value) {
      this.prompt = value;
      const { prompt, ...params } = this.modelValue || {};
      const own = !!(value || '').trim() && value !== (this.templateParams?.prompt ?? '');
      this.$emit('update:modelValue', own ? { ...params, prompt: value } : params);
    },
  },
};
</script>
