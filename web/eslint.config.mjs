// ESLint 9 flat config. The rule set is the one of the former .eslintrc.js: airbnb-base 15 (through
// @vue/eslint-config-airbnb), eslint-plugin-import recommended, eslint-plugin-vue `essential`
// (Vue 3) and eslint-plugin-vuejs-accessibility recommended, plus the project rules below.
import { createRequire } from 'node:module';
import { builtinRules } from 'eslint/use-at-your-own-risk';
import globals from 'globals';
import importPlugin from 'eslint-plugin-import';
import pluginVue from 'eslint-plugin-vue';
import pluginVueA11y from 'eslint-plugin-vuejs-accessibility';

const require = createRequire(import.meta.url);

// airbnb-base ships eslintrc files only; its rules are taken as they are. Rules that ESLint 9
// no longer has are dropped.
const airbnbRules = Object.assign({}, ...[
  require('eslint-config-airbnb-base/rules/best-practices'),
  require('eslint-config-airbnb-base/rules/errors'),
  require('eslint-config-airbnb-base/rules/node'),
  require('eslint-config-airbnb-base/rules/style'),
  require('eslint-config-airbnb-base/rules/variables'),
  require('eslint-config-airbnb-base/rules/es6'),
  require('eslint-config-airbnb-base/rules/imports'),
  require('eslint-config-airbnb-base/rules/strict'),
].map((config) => config.rules));

Object.keys(airbnbRules).forEach((rule) => {
  if (!rule.includes('/') && !builtinRules.has(rule)) {
    delete airbnbRules[rule];
  }
});

export default [
  {
    ignores: ['public/**', 'node_modules/**'],
  },

  // Sets ecmaVersion 2018: the language options below come after it.
  importPlugin.flatConfigs.recommended,

  {
    files: ['**/*.{js,mjs,cjs,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    linterOptions: {
      // ESLint 7 did not report unused disable directives.
      reportUnusedDisableDirectives: 'off',
    },
  },

  {
    rules: airbnbRules,
  },

  ...pluginVue.configs['flat/essential'],
  ...pluginVueA11y.configs['flat/recommended'],

  {
    settings: {
      'import/resolver': {
        node: {
          extensions: ['.js', '.vue', '.json'],
        },
        alias: {
          map: [['@', './src']],
          extensions: ['.vue', '.js'],
        },
      },
      'import/extensions': ['.js', '.jsx', '.mjs', '.ts', '.tsx'],
      'import/core-modules': [],
      'import/ignore': ['node_modules', '\\.(coffee|scss|css|less|hbs|svg|json)$'],
    },
    rules: {
      // @vue/eslint-config-airbnb
      'import/extensions': ['error', 'always', {
        js: 'never',
        mjs: 'never',
        jsx: 'never',
        ts: 'never',
        tsx: 'never',
      }],
      'vuejs-accessibility/no-onchange': 'off',
      'no-param-reassign': ['error', {
        props: true,
        ignorePropertyModificationsFor: ['state', 'acc', 'e'],
      }],

      // Defaults that changed in ESLint 9, set back to those of ESLint 7.
      'no-unused-vars': ['error', {
        vars: 'all', args: 'after-used', ignoreRestSiblings: true, caughtErrors: 'none',
      }],
      'no-constant-condition': ['warn', { checkLoops: true }],
      'no-inner-declarations': ['error', 'functions', { blockScopedFunctions: 'disallow' }],
      'no-useless-computed-key': ['error', { enforceForClassMembers: false }],

      // The project
      'no-bitwise': 'off',
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'linebreak-style': 'off',
      'prefer-destructuring': 'off',
      'vuejs-accessibility/click-events-have-key-events': 'off',
      'vuejs-accessibility/no-autofocus': 'off',
      // Added to the recommended set in eslint-plugin-vuejs-accessibility 2; not in the 1.x set.
      'vuejs-accessibility/no-static-element-interactions': 'off',
      'vue/valid-v-slot': 'off',
      'vue/multi-word-component-names': 'off',
      // Vue 3 removed these Vue 2 APIs: keep them out of the code.
      'vue/no-deprecated-events-api': 'error',
      'vue/no-deprecated-filter': 'error',
      'vue/no-deprecated-v-on-native-modifier': 'error',
      'vue/require-explicit-emits': 'error',
    },
  },

  {
    files: ['tests/unit/**/*.spec.js'],
    languageOptions: {
      globals: {
        ...globals.mocha,
        ...globals.vitest,
      },
    },
  },

  {
    // Build tooling: dependencies of the build only.
    files: ['vite.config.mjs', 'eslint.config.mjs'],
    rules: {
      'import/no-extraneous-dependencies': ['error', { devDependencies: true }],
    },
  },
];
