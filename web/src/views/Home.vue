<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="10" md="8" lg="6">
        <v-card class="pa-8 text-center" elevation="6">
          <div class="display-1 font-weight-bold mb-6">DIGITAL DATA FORGE</div>

          <div class="mb-8 subtitle-1">Build, Update, Comply</div>

          <v-btn color="primary" x-large @click="launch">
            LAUNCH APP
          </v-btn>

          <div class="mt-6 caption">
            digitaldata.co
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>

</template>

<script>
export default {
  name: 'Home',
  methods: {
    launch() {
      this.$router.push('/auth/login');
    },
  },
};
</script>

<style scoped>
</style>
