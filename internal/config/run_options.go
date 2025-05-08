package config

// RunOptions represents the configuration for a run command
type RunOptions struct {
	PortNumber   int    `json:"port_number"`
	IsDetached   bool   `json:"is_detached"`
	CommandToRun string `json:"command_to_run"`
	PRLink       string `json:"pr_link"`
	BranchName   string `json:"branch_name"`
	CommitHash   string `json:"commit_hash"`
	TagName      string `json:"tag_name"`
	NamedCommand string `json:"named_command"`
	RemoteName   string `json:"remote_name"`
}
