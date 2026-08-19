import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';
import OpenBaoIcon from '@/components/OpenBaoIcon.vue';
import DvlsIcon from '../components/DvlsIcon.vue';
import AwsSmIcon from '../components/AwsSmIcon.vue';
import AzureKvIcon from '../components/AzureKvIcon.vue';

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
      openbao: {
        component: OpenBaoIcon,
      },
      dvls: {
        component: DvlsIcon,
      },
      aws_sm: {
        component: AwsSmIcon,
      },
      azure_kv: {
        component: AzureKvIcon,
      },
    },
  },
});
