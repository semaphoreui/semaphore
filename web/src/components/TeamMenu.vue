<template>
  <v-tabs class="pl-4">
    <v-tab
      key="team"
      :to="`/project/${projectId}/team`"
      data-testid="team-members"
    >
      Members
    </v-tab>

    <v-tab
      v-if="systemInfo.teams.invites_enabled"
      key="invites"
      :to="`/project/${projectId}/invites`"
      data-testid="team-invites"
    >
      Invites
    </v-tab>

    <v-tab
      key="roles"
      :to="`/project/${projectId}/roles`"
      data-testid="team-roles"
    >
      Roles
    </v-tab>
  </v-tabs>
</template>

<script>
export default {

  props: {
    projectId: Number,
    systemInfo: Object,
  },

};
</script>
