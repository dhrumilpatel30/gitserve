# GitServe Module Structure

GitServe's architecture is designed for clarity and maintainability, organized into primary functional areas.

### Core Components

1. `cmd/` - Manages Command Line Interface operations.
   - Handles command parsing, validation, and CLI-specific error handling.
2. `internal/` - Contains the core business logic and services.
   - Houses all primary service implementations, detailed below.
3. `example/` - Provides a testing repository.
   - A simple NextJS startup template used for manual testing.

### Internal Services (within `internal/`)

The `internal/` directory is structured into the following packages. Each service package (e.g., `internal/git/`) implements a specific domain of logic for the application.

**Code Organization within a Service Package:**
To maintain readability and manageability as services grow, a service package can be composed of multiple `.go` files. For example, `internal/git/` might contain:

- `interface.go`: Defines the primary service interface (e.g., `git.Service`).
- `git.go` (or `service.go`): Contains the main struct implementing the interface (e.g., `git.ServiceImpl`) and its constructor (e.g., `NewService()`).
- Additional `ops_*.go` or feature-specific `.go` files (e.g., `ops_basic.go`, `prepare_repo.go`): These files contain the implementations of the service methods, logically grouped. All files within `internal/git/` would belong to `package git`.

This approach ensures that while a service like `git.Service` provides a cohesive API, its implementation details are spread across well-named files, preventing any single file from becoming overly large.

**Service Packages:**

1. `internal/git/` - Git Operations
2. `internal/instance/` - Instance and Process Management
3. `internal/storage/` - Storage
4. `internal/logger/` - Logging
5. `internal/port/` - Port Management
6. `internal/validation/` - Validation
7. `internal/config/` - Configuration Management
8. `internal/workspace/` - Temporary Directory Management
9. `internal/runner/` - Orchestrates running commands based on Git sources.
10. `internal/models/` - Contains shared data structures (requests, responses, core entities).
11. `internal/termui/` - Utilities for terminal output, like color codes.

## Command Flow

1. Run command
   `cmd/` → `runner/` (orchestrator) → `config/` → `git/` (uses its various internal files) → `workspace/` → `port/` → `instance/` → `storage/`
