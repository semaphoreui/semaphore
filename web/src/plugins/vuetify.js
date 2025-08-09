// Temporary Vuetify compatibility stub for Vue 3 migration stage
// Provides a minimal $vuetify global to avoid runtime errors
// eslint-disable-next-line import/extensions
import 'vuetify/styles';
import { createVuetify } from 'vuetify';
import { aliases, mdi } from 'vuetify/iconsets/mdi';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';

const vuetify = createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases: {
      ...aliases,
      tofu: { component: OpenTofuIcon },
      pulumi: { component: PulumiIcon },
      terragrunt: { component: TerragruntIcon },
      hashicorp_vault: { component: HashicorpVaultIcon },
    },
    sets: {
      mdi,
    },
  },
});

export default vuetify;
