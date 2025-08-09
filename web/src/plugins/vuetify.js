// Temporary Vuetify compatibility stub for Vue 3 migration stage
// Provides a minimal $vuetify global to avoid runtime errors
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';

export default {
  install(app) {
    const gp = app.config.globalProperties;
    gp.$vuetify = gp.$vuetify || {};
    const g = gp.$vuetify;
    g.theme = g.theme || { dark: false };
    g.icons = g.icons || { values: {} };
    g.icons.values = {
      ...g.icons.values,
      tofu: { component: OpenTofuIcon },
      pulumi: { component: PulumiIcon },
      terragrunt: { component: TerragruntIcon },
      hashicorp_vault: { component: HashicorpVaultIcon },
    };
  },
};
