### Service Platform

This project uses [Task](https://taskfile.dev/) as the build tool instead of Make.

#### Prerequisites

Install Task:

```bash
# macOS
brew install go-task/tap/go-task

# Or download from https://github.com/go-task/task/releases
```

#### Getting Started

```bash
task up-infra
task clean-migrate

task run
```

#### Available Tasks

Run `task --list` to see all available tasks:

- `task up-infra` - Start infrastructure services using Docker Compose
- `task down-infra` - Stop infrastructure services using Docker Compose
- `task migrate` - Run database migrations
- `task clean-migrate` - Clean and run database migrations for local database
- `task migrate-test` - Run database migrations for test database
- `task clean-migrate-test` - Clean and run database migrations for test database
- `task fmt` - Format Go code
- `task run` - Run the Go application
- `task integration-test` - Clean test database and run integration tests
- `task mockery` - Generate mocks for unit tests

#### Swagger generation

```bash
task swagger
// or
swag init -g app/cmd/api/main.go
```

Env

```env
APP_ENV=local || test || dev || production
```
