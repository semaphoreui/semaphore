import { test as base } from "@playwright/test";

export const test = base.extend<{
  login: (asAdmin: boolean) => Promise<void>;
  project: {
    create: (
      role?: "owner" | "manager" | "task_runner" | "guest",
      demo?: boolean,
      name?: string
    ) => Promise<string>;
    delete: () => Promise<void>;
  };
}>({
  login: async ({ page }, use) => {
    await use(async (asAdmin: boolean) => {
      await page.goto("/auth/login");
      const username = asAdmin
        ? process.env.TEST_ADMIN_LOGIN
        : process.env.TEST_USER_LOGIN;
      if (!username) {
        throw new Error("TEST_ADMIN_LOGIN or TEST_USER_LOGIN is not set");
      }
      const password = asAdmin
        ? process.env.TEST_ADMIN_PASSWORD
        : process.env.TEST_USER_PASSWORD;
      if (!password) {
        throw new Error("TEST_ADMIN_PASSWORD or TEST_USER_PASSWORD is not set");
      }
      await page.getByTestId("auth-username").fill(username);
      await page.getByTestId("auth-password").fill(password);
      await page.getByTestId("auth-signin").click();
    });
  },

  project: async ({ page }, use) => {
    await use({
      /**
       * Create a new project.
       * @param role   One of 'owner'|'manager'|'task_runner'|'guest'; defaults to 'owner'
       * @param demo   Whether to enable the "Demo" flag; defaults to false
       * @param name   Optional custom project name; if omitted, a timestamped name is generated
       * @returns      The name of the newly created project
       */
      create: async (role = "owner", demo = false, name?: string) => {
        const projectName = name ?? `test-${role}-${Date.now()}`;
        // open new-project dialog
        await page.getByTestId("sidebar-currentProject").click();
        await page.getByTestId("sidebar-newProject").click();
        // fill in details
        await page.getByTestId("newProject-name").fill(projectName);
        if (demo) {
          await page.getByRole("dialog").getByText("Demo").click();
        }
        // (optional) select role if your UI supports it:
        // await page.getByRole('combobox', { name: 'Role' }).selectOption(role);
        await page
          .getByRole("dialog")
          .getByRole("button", { name: "Create" })
          .click();

        await page.getByText(`Project ${projectName} created`).waitFor();

        // wait for the project to appear in the sidebar
        await page
          .getByTestId("sidebar-currentProject")
          .getByText(projectName)
          .waitFor();

        await page.waitForTimeout(500);

        return projectName;
      },

      /**
       * Delete an existing project by name.
       * @param name  The exact project name to delete
       */
      delete: async () => {

        await page.getByTestId("sidebar-dashboard").click();

        await page.getByTestId("dashboard-settings").click();

        await page.getByTestId("settings-deleteProject").click();

        await page
          .getByRole("dialog")
          .getByRole("button", { name: "Yes" })
          .click();
      },
    });
  },
});

export { expect } from "@playwright/test";
