---
name: semaphore:address-review
description: Evaluate and address supplied code review findings. Use when the user provides review comments to verify, prioritize, and fix based on test evidence.
---

# Address code review findings

1. Check each finding against the current code, diff, and relevant callers. Identify the expected behavior, the conditions that trigger the issue, and its concrete impact. Assess importance from that evidence. A regression in a supported workflow or a risk of data loss warrants a fix. Cosmetic issues and optional refactors usually have low impact. If a finding is incorrect or already resolved, explain the evidence and move on.
2. Proceed directly with confirmed important findings. For a confirmed low-impact finding, explain the impact, ask the user whether to fix it, and wait for approval before making related edits. Continue working on independent important findings while waiting.
3. Before changing the implementation for any fix, run a focused unit, integration, or end-to-end test on the current implementation. Reuse an existing test or add a regression test that exercises the affected behavior. Verify that the test fails for the reason described in the finding. If you cannot reproduce the issue, report the test result or blocker and defer that fix.
4. Make the smallest fix that addresses the verified issue. Rerun the same test and the relevant surrounding tests. In the summary, give the decision and evidence for each finding. For each fix, include the test command and its results before and after the change.
