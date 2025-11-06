import { test, expect } from "./fixtures";

test.describe("Variable Groups", () => {
  test.beforeEach(async ({ page, login, project }) => {
    await login(true);

    await project.create("manager", true);
  });

  test.afterEach(async ({ project, page }) => {
    await page
      .getByTestId("varGroupDialog")
      .getByTestId("editDialog-close")
      .click();
    await project.delete();
  });

  test("saving variables with empty names is prohibited", async ({
    page,
    login,
  }) => {
    await page.getByTestId("sidebar-environment").click();

    await page.getByRole("button", { name: "New Group" }).click();
    await page.getByRole("textbox", { name: "Group Name" }).fill("Test");

    await page.getByTestId("varGroup-addEnv").click();

    await page.getByRole("textbox", { name: "Value" }).fill("Test");
    await page.getByRole("button", { name: "Save" }).click();
    await page.getByTestId("varGroup-error").waitFor({ timeout: 1000 });
    await expect(page.getByTestId("varGroup-error")).toHaveText(
      "Environment variables key can not be empty"
    );
  });
});
