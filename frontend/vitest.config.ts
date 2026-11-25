import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    projects: ['packages/*', 'apps/*'],
    passWithNoTests: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      reportsDirectory: './coverage',
      include: ['packages/*/src/**/*.{ts,vue}', 'apps/*/src/**/*.{ts,vue}'],
      exclude: [
        'packages/forms-core/src/api/schema.d.ts',
        '**/*.d.ts',
        '**/*.spec.ts',
        '**/*.test.ts',
        '**/mocks/**',
        'apps/forms-admin/src/main.ts',
      ],
    },
  },
})
