---
name: semaphore-commit
---

Create well-formatted git commits following conventional commit standards.

## Behavior
1. Analyze staged changes with `git diff --staged`
2. Generate a conventional commit message
3. Create the commit with proper formatting
4. Do not add "Co-Authored-By: Claude ... <noreply@anthropic.com> to the footer"

## Commit Format
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Types
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Code style changes
- refactor: Code refactoring
- test: Adding or modifying tests
- chore: Maintenance tasks
- agents: Changes related to AI agents work

## Example Output
```
feat(auth): add password reset functionality

- Add forgot password form
- Implement email verification flow
- Add password reset endpoint
```
