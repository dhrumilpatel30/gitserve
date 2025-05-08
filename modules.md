# GitServe Module Structure

GitServe's architecture is designed for clarity and maintainability, organized into primary functional areas.

## Core Components

1.  `cmd/` - Manages Command Line Interface operations.
    - Handles command parsing, validation, and CLI-specific error handling.
2.  `internal/` - Contains the core business logic and services.
    - Houses all primary service implementations, detailed below.
3.  `example/` - Provides a testing repository.
    - A simple NextJS startup template used for manual testing.

## Internal Services (within `internal/`)

The `internal/` directory is structured into the following packages:

6.  `internal/git/` - Git Operations
7.  `internal/instance/` - Instance and Process Management
8.  `internal/storage/` - Storage
9.  `internal/logger/` - Logging
10. `internal/port/` - Port Management
11. `internal/validation/` - Domain Validation
12. `internal/config/` - Configuration Management
13. `internal/workspace/` - Temporary Directory Management
14. `internal/runner/` - running the commands

## Command Flow

1. Run command
   `cmd/` → `runner/` → `config/` → `git/` → `workspace/` → `port/` → `instance/` → `storage/`
