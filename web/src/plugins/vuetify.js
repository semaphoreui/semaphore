import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';

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
    },
  },
});
