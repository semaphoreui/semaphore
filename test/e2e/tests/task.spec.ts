import { test, expect } from "./fixtures";

test.describe("task running", () => {
  test.beforeEach(async ({ page, login, project }) => {
    await login(true);

    await project.create("task_runner", true);

    await page.getByTestId("sidebar-templates").click();

    await page.getByRole("link", { name: "Build demo app" }).click();

    await page.getByTestId("template-run").click();

    await page
      .getByTestId("newTaskDialog")
      .getByRole("textbox", { name: "Message (Optional)" })
      .fill("Test");

    await page
      .getByTestId("newTaskDialog")
      .getByTestId("editDialog-save")
      .click();

    test.setTimeout(90000);
  });

  test.afterEach(async ({ page, project }) => {
    await page
      .getByTestId("taskLogDialog")
      .getByTestId("editDialog-close")
      .click();

    await project.delete();
  });

  test("run task from demo project", async ({ page }) => {
    await page.getByTestId("task-rawLog").waitFor({ timeout: 90000 });

    await expect(page.getByTestId("task-status")).toHaveText("Success");
  });

  test("stop task on waiting", async ({ page }) => {
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Stop" })
      .click();

    await page.getByTestId("task-rawLog").waitFor({ timeout: 600000 });

    await expect(page.getByTestId("task-status")).toHaveText("Stopped");
  });

  test("stop task on cloning", async ({ page }) => {
    await page
      .getByRole("dialog")
      .getByText("Get current commit hash")
      .waitFor();

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Stop" })
      .click();

    await page.getByTestId("task-rawLog").waitFor({ timeout: 60000 });

    await expect(page.getByTestId("task-status")).toHaveText("Stopped");
  });

  test("stop task on running", async ({ page }) => {
    await page
      .getByRole("dialog")
      .getByText(
        "TASK [Gathering Facts] *********************************************************"
      )
      .waitFor({ timeout: 100000 });

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Stop" })
      .click();

    await page.getByTestId("task-rawLog").waitFor({ timeout: 60000 });

    await expect(page.getByTestId("task-status")).toHaveText("Stopped");
  });
});
