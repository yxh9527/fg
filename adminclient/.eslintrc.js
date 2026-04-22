module.exports = {
  root: true,
  'extends': [
    'plugin:vue/essential',
    '@vue/standard'
  ],
  rules: {
    // allow async-await
    'generator-star-spacing': 'off',
    'eqeqeq': 'off',
    'no-unused-vars': 'warn',
    'handle-callback-err': 'off',
    'no-useless-escape': 'off',
    'one-var': 'off',
    'standard/object-curly-even-spacing': 'off',
    'standard/computed-property-even-spacing': 'off',
    // allow debugger during development
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    'vue/no-parsing-error': [2, {
      'x-invalid-end-tag': false,
      'missing-semicolon-after-character-reference': false
    }],
    'no-undef': 'off',
    'camelcase': 'off'
  },
  parserOptions: {
    parser: 'babel-eslint'
  }
}
