## General Instructions

- Use Golang as the language for the development.
- Include a Makefile to define repetitive actions.
- When adding functionality, write unit tests for that functionality.
- All successful output from the utility should be in JSON format printed to STDOUT.
- Any errors should be in string format printed to STDERR.

## Documentation

- Use README.md to hold general documentation on the project.
- If needed additional Markdown files can be created in a "docs" sub-directory in the project.

## Dependencies

- Avoid introducing a new external dependencies unless absolutely necessary.
- If a new dependency is required, state the reason and request user approval.
- Always vendor dependencies in the "vendor" directory using `go mod vendor` to manage the vendoring of dependencies.

## Workflow

- Be sure to make sure the project builds successfully when you’re done making a series of code changes
- Prefer running single tests, and not the whole test suite, for performance
