import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

export default tseslint.config(
  {
    ignores: [
      '**/dist/**',
      '**/coverage/**',
      '**/public/mockServiceWorker.js',
      '**/src/api/schema.d.ts',
      '**/test-results/**',
      '**/playwright-report/**',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.strict,
  ...pluginVue.configs['flat/recommended'],
  {
    files: ['**/*.vue'],
    languageOptions: { parserOptions: { parser: tseslint.parser } },
  },
  {
    languageOptions: { globals: { ...globals.browser, ...globals.node } },
  },
  {
    // forms-core no depende de Vue (ver frontend/AGENTS.md).
    files: ['packages/forms-core/**'],
    rules: {
      'no-restricted-imports': ['error', { patterns: ['vue', 'vue/*', '@vue/*'] }],
    },
  },
  prettier,
)
