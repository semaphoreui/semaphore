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
    // `amber darken-3` of the palette with white text: Vuetify 2 drew coloured buttons and alerts
    // dark (white text), Vuetify 4 picks black for this colour.
    'amber-darken-3': '#FF8F00',
    'on-amber-darken-3': '#FFFFFF',
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
    'amber-darken-3': '#FF8F00',
    'on-amber-darken-3': '#FFFFFF',
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
    // Vuetify 4 follows the theme of the system by default; App.vue switches it.
    defaultTheme: 'light',
    // Variations of the theme colours, as Vuetify 2 generated them (`primary darken-2`, the same
    // CIELAB steps): lighten-1…5, darken-1…4.
    variations: {
      colors: ['primary', 'secondary', 'accent', 'error', 'info', 'success', 'warning'],
      lighten: 5,
      darken: 4,
    },
    themes: {
      // States of Vuetify 2 (`$material-light.states`, `$material-dark.states`): the dark theme
      // has twice the opacities of the light one; Vuetify 4 uses the light ones for both.
      light: {
        dark: false,
        colors: colors.light,
        variables: {
          'pressed-opacity': 0.16,
        },
      },
      dark: {
        dark: true,
        colors: colors.dark,
        variables: {
          'hover-opacity': 0.08,
          'focus-opacity': 0.24,
          'selected-opacity': 0.16,
          'activated-opacity': 0.24,
          'pressed-opacity': 0.32,
          'dragged-opacity': 0.16,
        },
      },
    },
  },
  // Defaults of Vuetify 2 that Vuetify 4 changed (the rest of the Vuetify 2 look is in
  // scss/variables.scss and assets/scss/vuetify2.scss).
  defaults: {
    // The selected tab and its slider in the primary colour.
    VTabs: { color: 'primary' },
    // A chip without a colour is grey, a coloured one is filled (Vuetify 4: tonal).
    VChip: { variant: 'flat' },
    VChipGroup: { variant: 'flat' },
    // Selection controls and focused fields in the primary colour (Vuetify 4: the text colour).
    VCheckbox: { color: 'primary' },
    VCheckboxBtn: { color: 'primary' },
    VSwitch: { color: 'primary', thumbColor: 'primary' },
    VRadioGroup: { color: 'primary' },
    VRadio: { color: 'primary' },
    // A field without a variant is underlined, as in Vuetify 2 (Vuetify 4: filled).
    VTextField: { color: 'primary', variant: 'underlined' },
    VTextarea: { color: 'primary', variant: 'underlined' },
    // The list of a select opens over the field, from its top edge (Vuetify 2 `v-select` menu
    // without `offset-y`; autocomplete and combobox open below, as there).
    VSelect: {
      color: 'primary',
      variant: 'underlined',
      menuProps: { location: 'top', origin: 'top' },
    },
    VAutocomplete: { color: 'primary', variant: 'underlined' },
    VCombobox: { color: 'primary', variant: 'underlined' },
    VFileInput: { color: 'primary', variant: 'underlined' },
    // Progress bars (and the loading bars of tables and fields) in the primary colour.
    VProgressLinear: { color: 'primary' },
    // The close button of an alert: a circled cross (`$cancel`).
    VAlert: { closeIcon: '$cancel' },
    VDataTable: {
      // Vuetify 2 defaults of `v-data-table`: rows as cards below 600 px (Vuetify 3 never
      // switches without a breakpoint) and the rows-per-page options of the footer.
      mobileBreakpoint: 600,
      itemsPerPageOptions: [5, 10, 15, -1],
    },
  },
  display: {
    // Breakpoints of Vuetify 2, not the smaller ones of Vuetify 4 (the same as $grid-breakpoints
    // in scss/variables.scss); `mobile` below lg, as in Vuetify 2.
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
