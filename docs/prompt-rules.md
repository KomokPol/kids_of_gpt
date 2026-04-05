# Prompt Rules for Agents

## Always include

- role of the service
- folder owner
- allowed files
- forbidden files
- contract version
- expected output

## Never say

- make it however you want
- change the contract if needed
- refactor adjacent services without asking

## Good prompt shape

1. Fix the task.
2. Fix the inputs.
3. Fix the outputs.
4. Fix the files that can change.
5. Fix the files that cannot change.
6. Ask only for the exact deliverable.