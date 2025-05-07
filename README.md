# GitServe
### After some chat with AI
Here's the list of things i want to ship:

### 1. Core Features

- **Run Commands from Git Source:**
  - `gitserve run <branch_name>`:
    - Uses local `<branch_name>` if it exists.
    - If not local, and `<branch_name>` has a configured upstream remote (e.g., `branch.branch_name.remote`), attempt to use that remote's version of `<branch_name>`.
    - If no local branch and no configured upstream for that specific branch name, an error will occur (unless `--remote` is specified).
  - `gitserve run <branch_name> --remote <remote_name>`: Explicitly run `<branch_name>` from `<remote_name>`.
  - `gitserve run --commit <commit_sha>`: Run from a specific commit.
  - `gitserve run --tag <tag_name>`: Run from a specific tag.
  - `gitserve run --pr <github_pr_url>`: Run code from a GitHub Pull Request.
  - `gitserve run --name <command_key>`: Execute a pre-defined command from `named_commands` in the config file, using the current Git context (default branch behavior or other source flags apply).
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

### 2. Config File Content (`.gitserve.toml`)

```toml
# Single command or array of commands to run before EACH main command.
# Think 'npm install', 'bundle install', etc.
pre_command = [
    "npm ci",
    "npm run build:icons"
]

# The go-to command if I just type 'gitserve run <branch_name>'
# and don't specify a named command.
default_run_command = "npm run dev"

# Port gitserve tries first.
default_port = 3000

# If default_port is taken, gitserve will try these in order.
preferred_ports_list = [3001, 3002, 8080, 8081, 5000]

# Handy: map certain branches to specific default ports.
# 'gitserve run main' would try 4000 first.
[branch_port_mapping]
main = 4000
develop = 4001
staging = 4002

# Your saved "recipes" for running things.
[named_commands.dev_server]
description = "Spins up the frontend dev server."
run_command = "npm run dev:frontend"
pre_command = "npm install --prefix frontend"  # Command-specific pre-command
env_vars = { NODE_ENV = "development", API_MOCK = "true" }

[named_commands.api_only]
description = "Runs just the backend API."
run_command = "npm run start:api"
default_port = 3005  # Command-specific default port
pre_command = [
    "npm install --prefix backend",
    "npm run migrate:dev --prefix backend"
]

[named_commands.test_suite]
description = "Runs all automated tests."
run_command = "npm test"

# Environment variables to apply to ALL commands gitserve runs.
# Not sure about the use case of this still let's have it.
[global_env_vars]
GITSERVE_MANAGED = "true"
LOG_LEVEL = "debug"
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


#### legacy things that i thought initially

## Things to do

1. Running the server from a branch(main feature)
2. Auto incrementing or changing the port if one is already in use.
3. Allowing running server on remote branch or PR directly? (still needs to brainstorm)
4. Again it is not just about the server we should be able to run any command per se.
5. Running detached servers with -d or --detach flags.
6. Allowing server management by list, stop, remove, stop-all commands.
7. allowing port rewrite if they want to with -p or --port flags.
8. help manual with -h or --help flag.
9. Interactive setup (required??) like if port already in use then ask, show default and allow to overwrite?
10. allow to run server on upstream branches by using some flag and specifying upstream like origin?
11. config file setup for projects(again optional but good to have).
12. init command to set up config file interactively
13. interactive terminal within that branch -i or —interactive with gitserve branch the -i tag will open child terminal with the functionality to run commands in it.

Config file contents

run_command, pre_command, default_port, named commands, which can also have set of commands too.

