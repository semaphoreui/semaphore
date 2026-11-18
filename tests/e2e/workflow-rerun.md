# Re-running a workflow

Use a fresh workflow for each pass through this suite. Run cases 1–5 in order. The letters A–F stand for run IDs recorded during the test. Task IDs and run IDs are separate. To read a task ID, click its node and record the number in the task log URL (`?t=` or `&t=`), then close the log.

## Fixture

Use the English UI on a Workflows-enabled stand with a local Bash runner. Sign in as the project owner. Use a workflow-enabled build of the PR in a private project. Keep task history for the duration of the suite and disable scheduled runs for this workflow.

1. Create a local repository with the two Bash scripts below and register it in the project. Ensure the local runner can read the repository. In a private temporary directory on the runner, create an empty file named `fail`. Substitute its absolute path for `/path/to/fail` below. Keep the scripts and templates unchanged throughout the suite.

   | Script | Body after `#!/bin/bash` |
   | --- | --- |
   | `ok.sh` | `exit 0` |
   | `probe.sh` | `test ! -f '/path/to/fail'` |

2. Create five Bash task templates named **Build**, **Probe**, **Publish**, **Recover** and **Audit**. Use the same repository and local inventory for all five. Select `probe.sh` for Probe and `ok.sh` for the other four.
3. Create a workflow named **Retry branches**, with start version `1.0.0`. Add one text survey variable named `release` with default value `default`. Add one node per template and an Approval node with the message **Allow probe?** and a one-hour timeout. Save the following graph as its first revision.

   | From | Edge condition | To |
   | --- | --- | --- |
   | Build | On success | Approval |
   | Approval | On success | Probe |
   | Approval | On success | Audit |
   | Probe | On success | Publish |
   | Probe | On failure | Recover |

4. Start one ordinary run with `release = qa-42`. Record its ID as **A**. Wait until **Allow probe?** appears. The **Re-run** button must be absent while approval is pending. Click **Approve** and wait for the run status to become `failed`.
5. Verify the baseline. Build, Audit and Recover succeeded. Probe ended in error. Publish shows **Not started**. Record the four task IDs, the revision displayed in the header and version `1.0.0`.
6. Edit the workflow, add a note reading **Revision two only**, and save. Confirm the note appears in the editor. Leave that revision as the current one.

Wait for the indicated UI state at each step. If a run does not reach it within two minutes on an idle local runner, record the current run status and the last task log as a failure.

## 1. Retry a failed run from the side panel while the fault persists

1. Leave the failure file in place. From **Workflows**, click A in the last-run column. In its side panel choose **Re-run → Only failed nodes**.
2. Record the new ID as **B**. The workflow list stays open and the panel URL contains `w=<workflow ID>/<B>`. The header links to **re-run of #A**, shows A's revision and version, and the variables show `release = qa-42`. The revision-two note is absent.
3. Wait for **Allow probe?**. Build and Audit already show **from run #A**. Open each and verify its task ID matches A. Probe has no task yet. Click **Approve**.
4. Wait for B to become `failed`. Probe fails with a new task ID. Recover succeeds with a new task ID even though it succeeded in A. Publish still shows **Not started**. Build and Audit retain their original task IDs.
5. Follow **re-run of #A**. The panel returns to A on the workflow list. A still shows its original failed status and four task IDs.

## 2. Retry the retry after fixing the fault

1. Remove the fixture's failure file from the runner. Use the browser Back button to return from A to B in the panel. On B choose **Re-run → Only failed nodes**. Record the new ID as **C**.
2. The header links to **re-run of #B**. Build and Audit both say **from run #A**. Click Build and verify its log opens with A's Build task ID.
3. Close the log and approve C. Wait for `success`. Probe has a new, successful task. Publish now has a successful task. Recover shows **Not started** even though it ran in both A and B. C retains version `1.0.0`, A's revision and `release = qa-42`.
4. Open **Re-run** on C. **Whole workflow** is enabled. **Only failed nodes** is disabled and displays **This run has no failed nodes**. Follow **re-run of #B** and check B still shows its failed Probe and successful Recover.

## 3. Stop a whole-workflow re-run at approval

1. Open C on its standalone page at `/project/<project ID>/workflows/<workflow ID>/runs/<C>`. Choose **Re-run → Whole workflow**. Record the new ID as **D**.
2. The browser stays on a standalone run page ending in `/runs/<D>`. The header links to C and shows version `1.0.1`. A's revision and `release = qa-42` are preserved. Build succeeds with a new task ID and no carry label.
3. Wait for **Allow probe?** and click **Stop** before approving. D reaches `stopped`. Probe, Publish, Recover and Audit have no tasks. Record D's Build task ID.

## 4. Resume the stopped work with a new approval

1. On D choose **Re-run → Only failed nodes**. Record the new ID as **E**. Wait for **Allow probe?**. The header shows **re-run of #D** and version `1.0.1`. Build shows **from run #D** and opens D's Build task. Probe and Audit have no tasks before approval.
2. Click **Approve**. E reaches `success`. Probe, Audit and Publish receive new successful tasks. Recover shows **Not started**.
3. Follow **re-run of #D**. The standalone URL ends in `/runs/<D>`. D is still stopped with only its Build task.

## 5. Re-run the whole graph from an older failed version

1. With the failure file still absent, open A's standalone page and choose **Re-run → Whole workflow**. Record the new ID as **F**. F links to A, uses A's revision and `release = qa-42`, and has version `1.0.2`. The revision-two note is absent.
2. Wait for **Allow probe?**, then approve. F reaches `success`. Build, Audit, Probe and Publish each have a new successful task ID. No card shows **from run**. Recover shows **Not started**.
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

Add a non-admin user with only the built-in Guest project role and no workflow-specific grants. Sign in as that user in a separate browser session.

1. Open A's standalone page. The run loads with status `failed` and the **Re-run** button is absent.
2. Using the guest session, send the POST from case 6 twice against A, once with mode `all` and once with mode `failed`. Both responses must be HTTP 403. Refresh history as the owner and verify no run was added.

## 8. Closing a loading panel stops polling its run

Use a browser request interceptor that can hold and release one response (for example, Playwright routing). Sign in as the owner, start a run and leave it waiting at Approval. Keep the **Workflows** list open with the run panel closed.

1. Intercept only the first GET to `/api/project/<project ID>/workflows/<workflow ID>/runs/<waiting run ID>`. Let the server respond and hold delivery of that response to the browser. Count all GETs to that exact path.
2. Open the waiting run from the list. Once the first response is held, close the panel using its close button. Verify the panel disappears, then release the held response and allow subsequent requests through.
3. Keep the tab in the foreground for 11 seconds (two five-second polling intervals). The GET count must remain **1** and the panel must stay closed. Reopen the run and verify its pending approval is visible. Stop it to finish the case.
