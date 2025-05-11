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

The `internal/` directory is structured into the following packages:

1. `internal/git/` - Git Operations
2. `internal/instance/` - Instance and Process Management
3. `internal/storage/` - Storage
4. `internal/logger/` - Logging
5. `internal/port/` - Port Management
6. `internal/validation/` - Validation
7. `internal/config/` - Configuration Management
8. `internal/workspace/` - Temporary Directory Management
9. `internal/runner/` - running the commands
10. `internal/models/` - contains basic request response models.

## Command Flow

1. Run command
   `cmd/` → `runner/` → `config/` → `git/` → `workspace/` → `port/` → `instance/` → `storage/`
