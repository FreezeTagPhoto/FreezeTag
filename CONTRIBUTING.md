# Contribution Guidelines/Pipeline
We want to keep things going smoothly while we work on the project, so we have to set a couple rules on how pull requests get merged.
## QA Testing
- As the author of a PR, you're expected to complete manual QA testing as well as implement unit tests for any new features.
- We require an 80% code coverage minimum for unit tests in the pipeline. If coverage drops below that point, the request cannot be merged until coverage is brought up.
## Review Process
- Repository maintainers will review your PR and may add comments. One approval from a maintainer is enough to be merged, but we will wait potentially several days for other maintainers to chime in.
- All change request threads must be resolved before the PR is merged, even if one maintainer approved it already.
## Automated Pipeline
- We will be running automated tests for the Go and NextJS portions of the codebase. If the coverage drops below 80% or any of the tests fail, this will block the merge until resolved.
- Even though the pipeline will only fail under 80% coverage, try to keep it as high as you can.
- The pipeline will also fail if formatting or lint checks fail. These are not skippable and should be resolved, however we may decide to change some lint rules if they happen to be too aggressive. 