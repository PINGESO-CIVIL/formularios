import { expect, test } from '@playwright/test'

test('la app carga', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Formularios')
})
