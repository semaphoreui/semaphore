import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';

Vue.use(Vuetify);

export default new Vuetify({
  icons: {
    values: {
      tofu: {
        component: OpenTofuIcon,
      },
      pulumi: {
        component: PulumiIcon,
      },
      terragrunt: {
        component: TerragruntIcon,
      },
      hashicorp_vault: {
        component: HashicorpVaultIcon,
      },
    },
  },
});
