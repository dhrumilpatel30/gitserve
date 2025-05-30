# GitServe

##### My vibe coded project, serious though. Readme Written After some chat with AI

Here's the list of things i want to ship:

### 1. Core Features

- **Run Commands from Git Source:**
  - `gitserve run <branch_name>`:
    - Uses local `<branch_name>` if it exists.
    - If not local then try to checkout the branch
  - `gitserve run <branch_name> --remote <remote_name>`: Explicitly run `<branch_name>` from `<remote_name>`.
  - `gitserve run --commit <commit_sha>`: Run from a specific commit.
  - `gitserve run --tag <tag_name>`: Run from a specific tag.
  - `gitserve run --pr <github_pr_url>`: Run code from a GitHub Pull Request.
- **Process Management:**
  - `-d, --detach`: Run the specified command in the background.
  - `list`: List all currently managed (running/detached) processes with ID, source, port, PID.
  - `stop <id>`: Stop a managed process by its ID (from `list`).
  - `logs <id>`: View logs of a detached process.
  - `remove <id>`: Stop and remove a managed process, cleaning up its temporary directory.
  - `stop-all`: Stop all managed processes.
- **Port Configuration:**
  - `-p, --port <port_number>`: Override the default port.
  - If a specified port is in use, the command will error out (simplification for now).
- **Usability:**
  - `-h, --help`: Display help manual for commands and subcommands.
  - `init`: Interactively create a `.gitserve.json` config file.
  - `-i, --interactive`: After cloning and running `pre_command`, open an interactive shell session within the temporary directory of the specified Git source.
- **Named Commands:**
  - `gitserve run --name dev_server` will run the sepcified things in the configuration.

### 2. Config File Content (`gitserve.yaml`)

```yaml
# Single command or array of commands to run before EACH main command.
# Think 'npm install', 'bundle install', etc.
pre_command:
  - npm ci
  - npm run build:icons

# The go-to command if I just type 'gitserve run <branch_name>'
# and don't specify a named command.
default_run_command: npm run dev

# Port gitserve tries first.
default_port: 3000

# If default_port is taken, gitserve will try these in order.
preferred_ports_list:
  - 3001
  - 3002
  - 8080
  - 8081
  - 5000

# Handy: map certain branches to specific default ports.
# 'gitserve run main' would try 4000 first.
branch_port_mapping:
  main: 4000
  develop: 4001
  staging: 4002

# Your saved "recipes" for running things.
named_commands:
  dev_server:
    description: Spins up the frontend dev server.
    run_command: npm run dev:frontend
    pre_command: npm install --prefix frontend # Command-specific pre-command
    env_vars:
      NODE_ENV: development
      API_MOCK: true

  api_only:
    description: Runs just the backend API.
    run_command: npm run start:api
    default_port: 3005 # Command-specific default port
    pre_command:
      - npm install --prefix backend
      - npm run migrate:dev --prefix backend

  test_suite:
    description: Runs all automated tests.
    run_command: npm test

# Environment variables to apply to ALL commands gitserve runs.
# Not sure about the use case of this still let's have it.
global_env_vars:
  GITSERVE_MANAGED: true
  LOG_LEVEL: debug
```

### 3. Architecture for Complex Features

- **GitHub PR (`--pr <url>`):**
  1.  Parse URL (owner, repo, PR number).
  2.  Utilize GitHub API (e.g., `google/go-github`) to fetch PR metadata (head branch, head repository clone URL).
      - Requires `GITHUB_TOKEN` (via environment variable) for private repositories or to avoid rate limiting. Fallback to unauthenticated for public repos.
  3.  Clone from the PR's head repository URL and head branch into a temporary directory.
  4.  We might need to support more providers like Gitlab, or bitbucket for but future.
- **Commit SHA (`--commit <sha>`):**
  1.  Clone the project repository (potentially a full clone, or a more optimized fetch of the specific commit if feasible) into a temporary directory.
  2.  Execute `git checkout <sha>` within that temporary directory.

### 4. Potential Improvements (Future Scope)

- **Workspace Caching/Re-use:**
  - Option to retain temporary directories for stopped (not removed) instances.
  - If re-running on an existing cached source, perform `git fetch` and `git reset --hard <ref>` instead of a full re-clone.
  - Configuration for cache size limits (number of instances or disk space).
- **Port Management:**
  - Auto-increment port if the default/specified one is in use, trying from a list or range.
  - Allow branch-to-port mapping in config for deterministic port assignment.
- **Instance Updates:** `gitserve update <id>` to fetch latest for the ref and restart the process.
- **Enhanced State:** Persist more detailed state about each instance for better management and re-use.
- **Branch Updates**: If the branch server is already running or it had run and directory is cached already then we might just need to update the dir and re run the server if already running.
