module.exports = {
  root: true,

  env: {
    node: true,
  },

  extends: [
    'plugin:import/recommended',
    'plugin:vue/essential',
    '@vue/airbnb',
  ],

  parserOptions: {
    parser: 'babel-eslint',
  },

  rules: {
    'no-bitwise': 'off',
    'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'linebreak-style': 'off',
    'prefer-destructuring': 'off',
    'vuejs-accessibility/click-events-have-key-events': 'off',
    'vue/valid-v-slot': 'off',
    'vue/multi-word-component-names': 'off',
  },

  overrides: [
    {
      files: [
        '**/__tests__/*.{j,t}s?(x)',
        '**/tests/unit/**/*.spec.{j,t}s?(x)',
      ],
      env: {
        mocha: true,
      },
    },
  ],

  settings: {
    'import/resolver': {
      node: {
        extensions: ['.js', '.vue'],
      },
      alias: {
        map: ['@', './src'],
        extensions: ['.vue', '.js'],
      },
    },
  },
};
