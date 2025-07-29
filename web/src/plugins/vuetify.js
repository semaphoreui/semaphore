import { createVuetify } from 'vuetify';
import 'vuetify/styles';
import { aliases, mdi } from 'vuetify/iconsets/mdi';
import '@mdi/font/css/materialdesignicons.css';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      mdi,
    },
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
  theme: {
    defaultTheme: 'light',
  },
});
