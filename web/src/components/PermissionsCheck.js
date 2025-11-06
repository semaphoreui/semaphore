export default {
  props: {
    userPermissions: Number,
    isAdmin: Boolean,
  },

  methods: {

    can(permission) {
      if (this.isAdmin) {
        return true;
      }

      if (this.item && this.item.permissions) {
        return (this.item.permissions & permission) === permission;
      }

      // eslint-disable-next-line no-bitwise
      return (this.userPermissions & permission) === permission;
    },
  },
};
