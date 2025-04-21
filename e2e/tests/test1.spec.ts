import { test, expect } from '@playwright/test';

test('test', async ({ page }) => {
  await page.goto('http://localhost:8080/auth/login');
  await page.getByTestId('auth-username').fill('fiftin');
  await page.getByTestId('auth-password').fill('150986');
  await page.getByTestId('auth-signin').click();

  await page.getByTestId('sidebar-templates').click();
  
  await page.getByRole('link', { name: 'Ping semaphoreui.com' }).click();
  await page.getByRole('button', { name: 'Run' }).click();
  await page.getByRole('textbox', { name: 'Message (Optional)' }).fill('Test');
  await page.getByRole('dialog').getByRole('button', { name: 'Run' }).click();

  await page.getByTestId('task-rawlog').waitFor({timeout: 100000});
});