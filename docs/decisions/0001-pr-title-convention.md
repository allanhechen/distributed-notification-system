# PR and Commit Title Convention

## Status

Accepted

## Context

As our Go project grows, we want to ensure that:

-   Pull requests (PRs) and commits are **clear and descriptive**.
-   We can easily generate **changelogs** or identify the type of change.
-   Code reviews are easier to follow.

Without a standard, contributors may use inconsistent titles like:

-   "Update stuff"
-   "Bug fix"
-   "New feature"

These titles make it hard to understand the purpose of the PR or commit at a glance.

## Decision

We will adopt a **structured PR and commit title convention** inspired by Conventional Commits:
<type>: <short description>

Where `<type>` is one of the following:

| Type     | When to use                              | Example                                   |
| -------- | ---------------------------------------- | ----------------------------------------- |
| feat     | New feature or functionality             | `feat: add user login endpoint`           |
| fix      | Bug fix                                  | `fix: prevent panic on nil input`         |
| docs     | Documentation only                       | `docs: update README with setup guide`    |
| chore    | Maintenance tasks (dependencies, config) | `chore: upgrade Go to version 1.21`       |
| refactor | Code restructuring (no behavior change)  | `refactor: simplify database connection`  |
| test     | Add or fix tests                         | `test: add unit tests for parser package` |
| perf     | Performance improvements                 | `perf: optimize JSON parsing`             |
| build    | Build system or CI config changes        | `build: add Go CI workflow`               |
| style    | Formatting, linting, whitespace          | `style: fix indentation in main.go`       |

### Examples

-   `feat: add new API endpoint for fetching users`
-   `fix: resolve crash when input is nil`
-   `docs: add example usage in README`
-   `chore: update dependencies`

## Consequences

-   PRs and commits will be **consistent** and easier to understand.
-   Changelogs and release notes can be **generated automatically**.
-   Contributors will need to **follow the convention**, which may require guidance for new team members.
