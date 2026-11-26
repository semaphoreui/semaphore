<template>
  <v-form
    ref="form"
    v-model="formValid"
    v-if="teamMembers != null && users != null"
    data-testid="teamMember-form"
    :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
      :model-value="!!formError"
      color="error"
      class="pb-2"
      data-testid="teamMember-error"
    >{{ formError }}
    </v-alert>

    <div class="d-flex justify-end mb-1" v-if="inviteType === 'both'">
      <v-btn-toggle
        v-model="selectedInviteType"
        rounded="0"
        variant="text"
        class="TeamMemberForm__inviteType"
        data-testid="teamMember-inviteType"
      >
        <v-btn
          value="email"
          size="small"
          class="mr-0"
          style="border-radius: 4px;"
          data-testid="teamMember-inviteType-email"
        >
          Email
        </v-btn>
        <v-btn
          value="username"
          size="small"
          class="mr-0"
          style="border-radius: 4px;"
          data-testid="teamMember-inviteType-username"
        >
          Username
        </v-btn>
      </v-btn-toggle>
    </div>

    <v-text-field
      v-if="selectedInviteType === 'email'"
      type="email"
      :label="$t('email')"
      v-model="item.email"
      data-testid="teamMember-email"
      variant="outlined"
      density="compact"
    />

    <v-autocomplete
      v-else
      v-model="item.user_id"
      data-testid="teamMember-userId"
      :label="$t('user')"
      :items="users"
      item-value="id"
      :item-title="(itm) => `${itm.username} (${itm.name})`"
      :rules="[v => !!v || $t('user_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      :menu-props="{ contentClass: 'TeamAutocompleteMenu' }"
    ></v-autocomplete>

    <v-select
      v-model="item.role"
      data-testid="teamMember-role"
      :label="$t('role')"
      :items="groupedRoles"
      item-value="slug"
      item-title="name"
      :rules="[v => !!v || $t('role_required')]"
      required
      :disabled="formSaving"
      variant="outlined"
      density="compact"
      :menu-props="{ location: 'top', origin: 'top', contentClass: 'TeamRoleMenu', maxHeight: 304 }"
    >
      <template v-slot:item="{ props, item: role }">
        <v-list-item
          v-bind="props"
          :title="role.name"
          :subtitle="role.slug"
          :data-testid="`teamMember-roleOption-${role.slug}`"
        />
      </template>
    </v-select>
  </v-form>
</template>
<style lang="scss">
// The list of an autocomplete has square corners (`.v-autocomplete__content.v-menu__content`).
.TeamAutocompleteMenu.v-overlay__content {
  border-radius: 0;

  > .v-sheet {
    border-radius: 0;
  }
}

// The role lists of the row field (Team.vue) and of this form: the dense list of a dense select of
// Vuetify 2 (`.v-list--dense`), at most 304 px high (`menuProps.maxHeight`) — group headers 40 px
// with the text 8 px in (`.v-subheader`, 12 px), two-line items 50 px: the name on a 16 px line,
// the slug 2 px under it at 13 px medium in the secondary text colour, also on the selected item
// (`.v-list-item__subtitle`).
.TeamRoleMenu {
  .v-list-subheader {
    min-height: 40px;
    padding-inline: 8px;
    font-size: 12px;
    line-height: 18px;
  }

  .v-list-item {
    min-height: 50px;
    padding-block: 8px;
  }

  .v-list-item-title {
    line-height: 16px;
  }

  .v-list-item-subtitle {
    margin-top: 2px;
    font-size: 13px;
    font-weight: 500;
    line-height: 16px;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    opacity: 1;
  }
}

// The Email / Username switch: a `group` toggle of Vuetify 2 (`.v-btn-toggle--group`) — small
// 28 px buttons 4 px apart inside a 36 px group, .8 opacity at rest; the active one in
// the secondary text colour on an .18 tint of it (`.v-btn--active::before`).
.TeamMemberForm__inviteType.v-btn-group {
  height: 36px;
}

.TeamMemberForm__inviteType > .v-btn {
  height: 28px;
  min-width: auto;
  margin: 4px 0 4px 4px;
  // The transparent thin frame of a toggle button, without the left side after the first
  // (`.v-btn-toggle > .v-btn.v-btn`).
  border: thin solid transparent;
  opacity: .8;

  &:not(:first-child) {
    border-left-width: 0;
  }

  &.v-btn--active {
    opacity: 1;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));

    > .v-btn__overlay {
      opacity: .18;
    }
  }
}
</style>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import groupRolesForSelect from '@/lib/roles';

export default {
  mixins: [ItemFormBase],

  props: {
    invitesEnabled: Boolean,
    inviteType: String,
    roles: Array,
  },

  computed: {
    availableRoles() {
      const roles = this.roles || [];
      if (!this.invitesEnabled) {
        return roles;
      }
      return roles.filter((role) => role.builtin);
    },

    groupedRoles() {
      return groupRolesForSelect(this.availableRoles, (key) => this.$t(key));
    },
  },

  data() {
    return {
      users: null,
      userId: null,
      teamMembers: null,
      selectedInviteType: this.inviteType === 'both' ? 'username' : this.inviteType,
    };
  },

  async created() {
    this.teamMembers = (await axios({
      method: 'get',
      url: this.getItemsUrl(),
      responseType: 'json',
    })).data;

    this.users = (await axios({
      method: 'get',
      url: '/api/users',
      responseType: 'json',
    })).data.filter((user) => !this.teamMembers.some((teamMember) => user.id === teamMember.id));
  },

  methods: {
    getItemsUrl() {
      if (this.invitesEnabled) {
        return `/api/project/${this.projectId}/invites`;
      }
      return `/api/project/${this.projectId}/users`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/users/${this.itemId}`;
    },
  },
};
</script>
