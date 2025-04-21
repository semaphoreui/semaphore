import { test, expect } from './fixtures';

test('run task from demo project', async ({ page, login, project }) => {
  await login(true);
  await project('task_runner');

  await page.getByTestId('sidebar-templates').click();
  
  await page.getByRole('link', { name: 'Ping semaphoreui.com' }).click();
  await page.getByRole('button', { name: 'Run' }).click();
  await page.getByRole('textbox', { name: 'Message (Optional)' }).fill('Test');
  await page.getByRole('dialog').getByRole('button', { name: 'Run' }).click();

  await page.getByTestId('task-rawlog').waitFor({timeout: 100000});
});