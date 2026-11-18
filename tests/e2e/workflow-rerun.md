# Re-running a workflow

Run cases **1–8 in order as one suite**, with a fresh private project and workflow for every pass. Cases 1–5 share the run chain A–F; case 6 needs A and C, case 7 needs A, and case 8 creates its own waiting run in the same workflow. Do not start a dependent case if its source run or expected state was not obtained. Record it as blocked and continue only with cases whose prerequisites are available.

The letters A–F stand for **run IDs**, not task IDs. Record both separately. To read a task ID, click its node, wait until the task log is visible **and** the URL contains `?t=<task ID>` or `&t=<task ID>`, and record that number. Close the log with its close button; wait until the log is hidden and `t` is absent from the URL before the next action. This also preserves the Browser Back check in case 2.

For **Not started / no task**, verify the node says **Not started** and cannot open a task log. A **new task ID** must differ from the IDs previously recorded for that template. A carry label **from run #X** must be checked separately from the task ID in the opened log.

## Stand and accounts

- Use the English UI on a prepared, idle test stand with Workflows enabled and the workflow re-run implementation under test. Record the application commit/build and the test-definition commit separately: a documentation-only test change does not require rebuilding the application. Do not start, restart or reconfigure the stand during this suite.
- Use the built-in local Bash executor: `use_remote_runner=false`, with no runner tag on the inventory or any template (`runner_tag=null`, not an empty string). An empty string can select remote-runner dispatch and is not the local-executor setting.
- The executor must be able to create, read and remove files in the local runner's filesystem. Run fixture shell/file operations as the runner's OS user. If the driver and runner are in different containers or filesystem namespaces, the stand must already provide an accessible private shared directory at the same absolute path. Without that access, record **BLOCKED**; do not change mounts or runner settings to make the test pass.
- Sign in using an instance administrator account, which becomes the owner of the fresh project it creates. Administrator access is needed to create the disposable non-admin user in case 7. Obtain its login/password from the test environment. If that account or those credentials are unavailable, record **BLOCKED** before creating the fixture.
- Use only objects created for this pass, with the `qa-` names below. Choose a unique suffix for the project name and Guest username, record the actual names, and use them consistently. Never reuse an older A–F chain or an unrelated user's account.

## Fixture

Create a private project named `qa-workflow-rerun-<unique suffix>` with the owner account: project selector → **New Project...** → **Project Name** → **Create**. Leave **Demo** off and the parallel-task limit at its default (no project limit). Keep all task/run history for the duration of the suite. The fresh workflow must have no schedules; do not disable schedules belonging to other workflows.

Project, key, repository, inventory, template, initial graph and Guest-account creation are fixture preparation and may use the UI or an authenticated HTTP client; record the method. The ordinary baseline run in fixture step 4, the revision edit in step 6 and the UI actions in cases 1–5 must use the UI. Use HTTP for the checks explicitly specified in cases 6–7 and a browser request interceptor for case 8.

Use these exact fixture settings; leave optional arguments, variable groups and other overrides empty/default:

| Object | Settings |
| --- | --- |
| Access key | Name `qa-none`; type **None**; no secret material. |
| Repository | Name `qa-local-workflow-scripts`; **local** directory `<fixture-dir>/repo`, not a `file://` Git clone; empty branch; access key `qa-none`. No `.git` directory is required for a local repository. |
| Inventory | Name `qa-local`; type **Static**; content `localhost ansible_connection=local`; user credentials `qa-none`; no privilege-escalation credentials; no repository override; no runner tag (`null`). |
| All five templates | App **Bash**; task type **Task** (including the template named `qa-Build`); the same repository and inventory above; **Allow parallel tasks** enabled; no runner tag (`null`); no arguments, variable groups, survey or version override. |
| Workflow | Name `qa-Retry branches`; start version `1.0.0`; one text survey variable with name/title `release` and default `default`; no schedules. |

1. In the runner's filesystem, create a fresh private temporary directory `<fixture-dir>` with a `qa-workflow-rerun-` prefix and mode `0700`. Create its `repo` subdirectory with mode `0700`. Create an empty file `<fixture-dir>/fail` with mode `0600`. Create the two Bash scripts below in `repo`, each beginning with `#!/bin/bash`, with mode `0700`. Substitute the absolute path `<fixture-dir>/fail` for `/path/to/fail`, and register the directory as the local repository specified above. Confirm that the runner's OS user can read both scripts and the failure file. Keep the script contents and templates unchanged throughout the suite; only remove the failure file when case 2 explicitly requests it.

   | Script | Body after `#!/bin/bash` |
   | --- | --- |
   | `ok.sh` | `exit 0` |
   | `probe.sh` | `test ! -f '/path/to/fail'` |

2. Create five Bash task templates named **qa-Build**, **qa-Probe**, **qa-Publish**, **qa-Recover** and **qa-Audit**. Use the same repository and local inventory for all five. Select `probe.sh` for qa-Probe and `ok.sh` for the other four.
3. Create a workflow named **qa-Retry branches**, with start version `1.0.0`. Add one text survey variable named `release` with default value `default`. Add one node per template and an Approval node with the message **Allow probe?** and a one-hour timeout. Save the following graph as its first revision.

   | From | Edge condition | To |
   | --- | --- | --- |
   | qa-Build | On success | Approval |
   | Approval | On success | qa-Probe |
   | Approval | On success | qa-Audit |
   | qa-Probe | On success | qa-Publish |
   | qa-Probe | On failure | qa-Recover |

4. Start one ordinary run with `release = qa-42`. Record its ID as **A**. Wait until **Allow probe?** appears. The **Re-run** button must be absent while approval is pending. Click **Approve** and wait for the run status to become `failed`.
5. Verify the baseline. qa-Build, qa-Audit and qa-Recover succeeded. qa-Probe ended in error. qa-Publish shows **Not started**. Record the four task IDs, the revision displayed in the header and version `1.0.0`.
6. Edit the workflow, add a note reading **Revision two only**, and save. Confirm the note appears in the editor. Leave that revision as the current one.

Wait for the indicated UI state at each step before reading IDs or advancing; a URL or dialog from the previous action is not sufficient. For run progress after Run, Re-run, Approve or Stop, allow up to two minutes on the idle local runner. Approval waiting for the tester's decision is intentional; begin the completion timeout after Approve or Stop. If the indicated progress state is not reached, record the current run status and the latest available task log as a failure; do not retry the operation or change the fixture to hide the failure. Unavailable stand/account/filesystem prerequisites are **BLOCKED**, not an application failure.

## 1. Retry a failed run from the side panel while the fault persists

1. Leave the failure file in place. From **Workflows**, click A in the last-run column. In its side panel choose **Re-run → Only failed nodes**.
2. Record the new ID as **B**. The workflow list stays open and the panel URL contains `w=<workflow ID>/<B>`. The header links to **re-run of #A**, shows A's revision and version, and the variables show `release = qa-42`. The revision-two note is absent.
3. Wait for **Allow probe?**. qa-Build and qa-Audit already show **from run #A**. Open each and verify its task ID matches A. qa-Probe has no task yet. Click **Approve**.
4. Wait for B to become `failed`. qa-Probe fails with a new task ID. qa-Recover succeeds with a new task ID even though it succeeded in A. qa-Publish still shows **Not started**. qa-Build and qa-Audit retain their original task IDs.
5. Follow **re-run of #A**. The panel returns to A on the workflow list. A still shows its original failed status and four task IDs.

## 2. Retry the retry after fixing the fault

1. Remove the fixture's failure file from the runner. Use the browser Back button to return from A to B in the panel. On B choose **Re-run → Only failed nodes**. Record the new ID as **C**.
2. The header links to **re-run of #B**. qa-Build and qa-Audit both say **from run #A**. Click qa-Build and verify its log opens with A's qa-Build task ID.
3. Close the log and approve C. Wait for `success`. qa-Probe has a new, successful task. qa-Publish now has a successful task. qa-Recover shows **Not started** even though it ran in both A and B. C retains version `1.0.0`, A's revision and `release = qa-42`.
4. Open **Re-run** on C. **Whole workflow** is enabled. **Only failed nodes** is disabled and displays **This run has no failed nodes**. Follow **re-run of #B** and check B still shows its failed qa-Probe and successful qa-Recover.

## 3. Stop a whole-workflow re-run at approval

1. Open C on its standalone page at `/project/<project ID>/workflows/<workflow ID>/runs/<C>`. Choose **Re-run → Whole workflow**. Record the new ID as **D**.
2. The browser stays on a standalone run page ending in `/runs/<D>`. The header links to C and shows version `1.0.1`. A's revision and `release = qa-42` are preserved. qa-Build succeeds with a new task ID and no carry label.
3. Wait for **Allow probe?** and click **Stop** before approving. D reaches `stopped`. qa-Probe, qa-Publish, qa-Recover and qa-Audit have no tasks. Record D's qa-Build task ID.

## 4. Resume the stopped work with a new approval

1. On D choose **Re-run → Only failed nodes**. Record the new ID as **E**. Wait for **Allow probe?**. The header shows **re-run of #D** and version `1.0.1`. qa-Build shows **from run #D** and opens D's qa-Build task. qa-Probe and qa-Audit have no tasks before approval.
2. Click **Approve**. E reaches `success`. qa-Probe, qa-Audit and qa-Publish receive new successful tasks. qa-Recover shows **Not started**.
3. Follow **re-run of #D**. The standalone URL ends in `/runs/<D>`. D is still stopped with only its qa-Build task.

## 5. Re-run the whole graph from an older failed version

1. With the failure file still absent, open A's standalone page and choose **Re-run → Whole workflow**. Record the new ID as **F**. F links to A, uses A's revision and `release = qa-42`, and has version `1.0.2`. The revision-two note is absent.
2. Wait for **Allow probe?**, then approve. F reaches `success`. qa-Build, qa-Audit, qa-Probe and qa-Publish each have a new successful task ID. No card shows **from run**. qa-Recover shows **Not started**.
3. Follow **re-run of #A**. A still has version `1.0.0`, its original task IDs and status `failed`.

## 6. Reject invalid requests without creating runs

As the owner, use an authenticated HTTP client against the same stand. For the first row, start a separate ordinary run and leave it waiting at Approval. Before each request record the set of IDs in the workflow run history. POST the JSON body to `/api/project/<project ID>/workflows/<workflow ID>/runs/<source ID>/rerun` with content type `application/json`.

| Source | Body | Expected HTTP status | Expected JSON `error` |
| --- | --- | --- | --- |
| Run waiting at Approval | `{"mode":"all"}` | 400 | `workflow run is still in progress` |
| Successful C | `{"mode":"failed"}` | 400 | `workflow run has no failed nodes to re-run` |
| Failed A | `{"mode":"invalid"}` | 400 | `workflow re-run mode is invalid` |

Refresh the run history after each response. The set of run IDs must match the set recorded before that request. Stop the waiting run after completing the checks.

## 7. A read-only member can inspect a failure

As the administrator/owner account used for the fixture, create a fresh non-admin user:

| Field | Value |
| --- | --- |
| Username | `qa-workflow-guest-<unique suffix>` |
| Display name | `qa-workflow-guest` |
| Email | `<username>@example.invalid` (test-only address; do not send mail) |
| Password | Generate a disposable random password with at least 128 bits of entropy when creating the user; keep it in memory or the test environment, not in the repository/report. |
| Account flags | Admin off; external authentication off; Pro off. |
| Project membership | Add to this pass's private project with **only** the built-in **Guest** role; verify the assigned-role list contains only Guest. |
| Workflow grants | No workflow-specific grants for this user. |

Record the run-history IDs as the owner before the requests. Sign in as the new user through the login form in a separate browser context/session with its own cookies; keep the owner's session open for the history check. If user creation or role assignment is unavailable, record **BLOCKED** rather than substituting another account.

1. Open A's standalone page. The run loads with status `failed` and the **Re-run** button is absent.
2. Using the guest session, send the POST from case 6 twice against A, once with mode `all` and once with mode `failed`. Both responses must be HTTP 403. Refresh history as the owner and verify no run was added.

## 8. Closing a loading panel stops polling its run

Use a browser request interceptor that can fetch the server response and delay only its delivery to the page (for example, Playwright `route.fetch()` followed by deferred `route.fulfill()`). Do not replace the response body, freeze the browser clock or delay unrelated requests. Use the default five-second run-polling interval. In headless mode, bring the page to the front and verify `document.visibilityState` is `visible` for the observation window.

Use the existing owner session. Start a separate ordinary run, record its ID, and leave it waiting at **Allow probe?** without approving. Close its panel and keep the **Workflows** list open. Do not install the interceptor until that panel is closed; earlier requests used to prepare the waiting run are outside the case's counter.

1. Intercept only the first GET to `/api/project/<project ID>/workflows/<workflow ID>/runs/<waiting run ID>`. Let the server respond and hold delivery of that response to the browser. Count all GETs to that exact path.
2. Open the waiting run from the list. Once the first response is held, close the panel using its close button. Verify the panel disappears, then release the held response and allow subsequent requests through.
3. Keep the tab in the foreground for 11 seconds (two five-second polling intervals). The GET count must remain **1** and the panel must stay closed. Reopen the run and verify its pending approval is visible. Stop it to finish the case.

## Evidence and cleanup

Record the application build, test-definition commit, project/workflow IDs, actual fixture names, absolute fixture directory, A–F run IDs, task IDs, revisions and versions. On an unexpected result, save a screenshot before cleanup and record expected versus actual values; do not replace an expected value with observed behavior.

After recording evidence, stop any still-active runs created by this pass so that they are not left waiting for approval. Touch only this pass's `qa-` fixtures. Preserve its run/task history and record which fixture directory and objects remain for inspection; do not delete shared data or reconfigure the stand. Cases 6 and 8 must still perform their own explicit Stop checks before this final cleanup.
