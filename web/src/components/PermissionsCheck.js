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
      // eslint-disable-next-line no-bitwise
      return (this.userPermissions & permission) === permission;
    },
  },
};
