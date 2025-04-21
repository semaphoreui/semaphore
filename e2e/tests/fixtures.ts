import { test as base } from '@playwright/test';

export const test = base.extend<{
  login: (asAdmin: boolean) => Promise<void>;
  project: (role: 'owner' | 'manager' | 'task_runner' | 'guest') => Promise<void>;
}>({
    
  login: async ({ page }, use) => {
    await use(async (asAdmin: boolean) => {
        await page.goto('/auth/login');
        const username = asAdmin ? process.env.TEST_ADMIN_LOGIN : process.env.TEST_USER_LOGIN;
        if (!username) {
            throw new Error('TEST_ADMIN_LOGIN or TEST_USER_LOGIN is not set');
        }
        const password = asAdmin ? process.env.TEST_ADMIN_PASSWORD : process.env.TEST_USER_PASSWORD;
        if (!password) {
            throw new Error('TEST_ADMIN_PASSWORD or TEST_USER_PASSWORD is not set');
        }
        await page.getByTestId('auth-username').fill(username);
        await page.getByTestId('auth-password').fill(password);
        await page.getByTestId('auth-signin').click();
    });
  },

  project: async ({ page }, use) => {
    await use(async (role = 'owner') => {
        await page.getByTestId('sidebar-currentProject').click();
        await page.getByTestId('sidebar-newProject').click();
        await page.getByRole('textbox', { name: 'Project Name' }).fill('Test');
        await page.getByRole('button', { name: 'Create' }).click();
    });
  }
});

export { expect } from '@playwright/test';