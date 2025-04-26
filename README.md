# ff-merge

ff-merge lets you do `git merge --ff-only` on GitHub.

## Supported Events

ff-merge supports events that (1) contains "issue" field with non-empty "pull_request" field, (2) contains "pull_request" field.

Notable events:

- [`pull_request`](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#pull_request): See[`labeled` example](https://github.com/oneee-playground/gh-fast-forward-test/blob/main/.github/workflows/ff-label.yml).
- [`issue_comment`](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#issue_comment): Only when the comment is from PR. See [example](https://github.com/oneee-playground/gh-fast-forward-test/blob/main/.github/workflows/ff-comment.yml).

## Inputs

| Name          | Type             | Choices                                                                  | Default               | Description                                                                                   |
| ------------- | ---------------- | ------------------------------------------------------------------------ | --------------------- | --------------------------------------------------------------------------------------------- |
| merge         | `string`         | `"true"`, `"false"`                                                      | `"false"`             | Whether to merge the branch immediately.                                                      |
| comment       | `string`         | `"never"`, `"failure"`, `"always"`                                       | `"always"`            | Determines if it should leave a comment.                                                      |
| allowed-roles | `list of string` | `"read"`, `"triage"`, `"write"`, `"maintain"`, `"admin"`, `$custom_role` | `[]`                  | Required role(s) for `triggering_actor` to trigger fast-forward. Example: `["read", "write"]` |
| GITHUB_TOKEN  | `string`         | Your PAT or auto-provided token                                          | `${{ github.token }}` | Token used to authenticate                                                                    |

## Example

See [oneee-playground/gh-fast-forward-test](https://github.com/oneee-playground/gh-fast-forward-test).
