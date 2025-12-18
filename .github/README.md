# GitHub Configuration for Mock Exam Project

## Workflows

This directory contains GitHub Actions workflows that automate testing, building, and quality checks for the Mock Exam project.

### CI Workflow

The continuous integration workflow (`ci.yml`) includes:
- Go version compatibility testing (1.21.x, 1.22.x)
- Unit tests and code validation
- Security scanning with govulncheck
- Code quality checks with golangci-lint
- Docker image building verification