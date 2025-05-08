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

4.  `internal/config/` - Configuration Management
    - Parses and validates configurations, handles environment variables, and defines defaults.
5.  `internal/model/` - Data Models
    - Defines common data structures (Request/Response types) and shared types (IDs, enums).
6.  `internal/git/` - Git Operations
    - Manages all Git-related tasks, including repository interaction, branch and PR handling.
7.  `internal/instance/` - Instance and Process Management
    - Oversees application instance lifecycles, state (persistence, synchronization, recovery), process control (signal handling, shutdowns), and health monitoring. (Consolidates previous `instance`, `state`, and `process` modules).
8.  `internal/storage/` - Storage and Caching
    - Provides the storage layer for Git repositories, manages file system operations, temporary directories, and workspace caching. (Consolidates previous `storage` and `cache` modules).
9.  `internal/logger/` - Logging
    - Implements structured logging, log level management, and log rotation.
10. `internal/port/` - Port Management
    - Handles port availability checking, allocation strategies, and conflict resolution.
11. `internal/errors/` - Error Handling
    - Defines custom error types, provides error handling utilities, and integrates with logging.
12. `internal/validation/` - Domain Validation
    - Contains domain-level validation rules, utilities, and custom validators.

## Testing

13. `test/` - Test Suites
    - Contains unit tests, integration tests, test fixtures, and mock implementations.
14. `internal/testutils/` - Test Utilities
    - Provides testing utilities, mock generators, and common test setup/teardown helpers.

## Module Interactions and Design Decisions

1.  **Interface Placement**:
    - Each module defines its own interfaces (e.g., `instance.Manager`, `storage.Repository`).
    - Common types and models reside in the `model` package.
2.  **Validation Strategy**:
    - `cmd` layer: Handles flag validation and basic format checking.
    - `validation` package: Implements complex domain-specific rules.
    - Each module: Contains self-contained business rules validation.
3.  **State Management**:
    - Centralized within the `instance` module, featuring event-driven updates and persistent storage integration.
4.  **Error Handling**:
    - Domain-specific errors are defined in their respective modules.
    - Common error types and utilities are located in the `errors` package.
    - Error wrapping with contextual information is encouraged.
5.  **Dependencies**:
    ```
    cmd → validation → internal services (e.g., git, instance, storage) → model
    ```
