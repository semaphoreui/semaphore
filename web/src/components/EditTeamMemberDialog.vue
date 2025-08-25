<template>
  <EditDialog
    v-model="dialog"
    :save-button-text="(
      itemId === 'new'
        ? (invitesEnabled ? 'Invite' : 'Link')
        : $t('save')
    )"
    :title="$t('teamMember', { expr: itemId === 'new' ? $t('nnew') : $t('edit') })"
    @save="onSave"
  >
    <template v-slot:form="{ onSave, onError, needSave, needReset }">
      <TeamMemberForm
        :project-id="projectId"
        :item-id="itemId"
        @save="onSave"
        @error="onError"
        :need-save="needSave"
        :need-reset="needReset"
        :invites-enabled="invitesEnabled"
        :invite-type="inviteType"
      />
    </template>
  </EditDialog>

</template>

<style lang="scss">
</style>

<script>
import EditDialog from './EditDialog.vue';
import TeamMemberForm from './TeamMemberForm.vue';

export default {
  components: {
    EditDialog,
    TeamMemberForm,
  },

  props: {
    value: Boolean,
    projectId: Number,
    itemId: [String, Number],
    invitesEnabled: Boolean,
    inviteType: String,
  },

  data() {
    return {
      dialog: false,
    };
  },

  watch: {
    async dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      this.dialog = val;
    },
  },

  methods: {
    onSave(e) {
      this.$emit('save', e);
    },
  },
};
</script>
