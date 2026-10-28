export default {
  props: {
    userPermissions: Number,
    isAdmin: Boolean,
  },

  methods: {

    // The object's permissions come from the server: the project verbs plus the grants on it.
    can(permission, item = this.item) {
      if (this.isAdmin) {
        return true;
      }

      if (item && item.permissions != null) {
        return (item.permissions & permission) === permission;
      }

      // eslint-disable-next-line no-bitwise
      return (this.userPermissions & permission) === permission;
    },
  },
};
