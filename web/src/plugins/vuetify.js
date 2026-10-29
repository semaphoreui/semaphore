import { defineComponent, h } from 'vue';
import { createVuetify } from 'vuetify';
import { aliases, mdi } from 'vuetify/iconsets/mdi';
import OpenTofuIcon from '@/components/OpenTofuIcon.vue';
import PulumiIcon from '@/components/PulumiIcon.vue';
import TerragruntIcon from '@/components/TerragruntIcon.vue';
import HashicorpVaultIcon from '@/components/HashicorpVaultIcon.vue';
import OpenBaoIcon from '@/components/OpenBaoIcon.vue';
import DvlsIcon from '../components/DvlsIcon.vue';
import AwsSmIcon from '../components/AwsSmIcon.vue';
import AzureKvIcon from '../components/AzureKvIcon.vue';

// Own SVG icons: the icon set `semaphore` (`semaphore:<name>`) and the aliases `$<name>`.
const SEMAPHORE_ICONS = {
  tofu: OpenTofuIcon,
  pulumi: PulumiIcon,
  terragrunt: TerragruntIcon,
  hashicorp_vault: HashicorpVaultIcon,
  openbao: OpenBaoIcon,
  dvls: DvlsIcon,
  aws_sm: AwsSmIcon,
  azure_kv: AzureKvIcon,
};

// Renders the SVG component inside the tag of `v-icon`, as Vuetify does for a component icon.
const SemaphoreIcon = defineComponent({
  name: 'SemaphoreIcon',
  props: {
    icon: [String, Object, Function, Array],
    tag: {
      type: String,
      required: true,
    },
  },
  render() {
    const Icon = SEMAPHORE_ICONS[this.icon];
    return h(this.tag, Icon ? [h(Icon)] : []);
  },
});

const semaphoreAliases = Object.fromEntries(
  Object.keys(SEMAPHORE_ICONS).map((name) => [name, `semaphore:${name}`]),
);

// Default palette of Vuetify 2.
const colors = {
  light: {
    primary: '#1976D2',
    secondary: '#424242',
    accent: '#82B1FF',
    error: '#FF5252',
    info: '#2196F3',
    success: '#4CAF50',
    warning: '#FB8C00',
  },
  dark: {
    background: '#1E1E1E',
    primary: '#2196F3',
    secondary: '#424242',
    accent: '#FF4081',
    error: '#FF5252',
    info: '#2196F3',
    success: '#4CAF50',
    warning: '#FB8C00',
  },
};

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases: {
      ...aliases,
      ...semaphoreAliases,
    },
    sets: {
      mdi,
      semaphore: { component: SemaphoreIcon },
    },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: { dark: false, colors: colors.light },
      dark: { dark: true, colors: colors.dark },
    },
  },
  defaults: {
    VDataTable: {
      // Vuetify 2 defaults of `v-data-table`: rows as cards below 600 px (Vuetify 3 never
      // switches without a breakpoint) and the rows-per-page options of the footer.
      mobileBreakpoint: 600,
      itemsPerPageOptions: [5, 10, 15, -1],
    },
  },
  display: {
    // Breakpoints of Vuetify 2 (the same as $grid-breakpoints in scss/variables.scss);
    // `mobile` below lg, as in Vuetify 2.
    mobileBreakpoint: 'lg',
    thresholds: {
      xs: 0,
      sm: 600,
      md: 960,
      lg: 1264,
      xl: 1904,
      xxl: 2560,
    },
  },
});
