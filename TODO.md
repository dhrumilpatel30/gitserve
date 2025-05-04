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
