package cmd

import (
	"errors"
	"fmt"
	"time"

	// "gitserve/internal/instance" // Not strictly needed if we signal directly
	"gitserve/internal/storage"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	colorResetStop  = "\033[0m"
	colorRedStop    = "\033[31m"
	colorGreenStop  = "\033[32m"
	colorYellowStop = "\033[33m"
	colorGrayStop   = "\033[90m"
	colorBoldStop   = "\033[1m"
)

var stopCmd = &cobra.Command{
	Use:   "stop [INSTANCE_ID]",
	Short: "Stop a running gitserve instance",
	Long:  `Stops a specific gitserve instance by its ID. The instance must be in a 'running' state.`,
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the instance ID
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceID := args[0]

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		storeDataPath := filepath.Join(homeDir, ".gitserve", "store")

		instanceStore, err := storage.NewJSONInstanceStore(storeDataPath)
		if err != nil {
			return fmt.Errorf("failed to initialize instance store: %w", err)
		}

		storedInst, found, err := instanceStore.GetInstanceByID(instanceID)
		if err != nil {
			return fmt.Errorf("failed to retrieve instance '%s%s%s': %w", colorBoldStop, instanceID, colorResetStop, err)
		}
		if !found {
			return fmt.Errorf("no instance found with ID '%s%s%s'", colorBoldStop, instanceID, colorResetStop)
		}

		if storedInst.Status != "running" {
			return fmt.Errorf("instance '%s%s%s' is not in a '%srunning%s' state (current status: %s%s%s). Cannot stop.",
				colorBoldStop, instanceID, colorResetStop,
				colorGreenStop, colorResetStop,
				colorYellowStop, storedInst.Status, colorResetStop)
		}

		if storedInst.PID == 0 {
			// originalStatus := storedInst.Status // Unused in this block
			storedInst.Status = "error_pid_zero"
			storedInst.StopTime = time.Now().UTC()
			if updateErr := instanceStore.UpdateInstance(instanceID, storedInst); updateErr != nil {
				cmd.PrintErrf("%sAdditionally, failed to update instance status to '%serror_pid_zero%s': %v%s\n", colorRedStop, colorRedStop, colorResetStop, updateErr, colorResetStop)
			}
			return fmt.Errorf("instance '%s%s%s' has PID 0 recorded, cannot stop. Status updated to '%serror_pid_zero%s'.",
				colorBoldStop, instanceID, colorResetStop, colorRedStop, colorResetStop)
		}

		fmt.Printf("Attempting to stop instance '%s%s%s' (PGID to be used: %s%d%s)...\n",
			colorBoldStop, storedInst.ID, colorResetStop, colorBoldStop, storedInst.PID, colorResetStop)

		// We don't need to os.FindProcess anymore if we are killing by PGID directly using syscall.Kill
		// proc, err := os.FindProcess(storedInst.PID)
		// if err != nil {
		// 	// Error finding process (e.g., os.ErrProcessDone if already exited)
		// 	fmt.Printf("Failed to find process with PID %d for instance '%s': %v. It might have already exited.\n", storedInst.PID, instanceID, err)
		// 	// Update status in store
		// 	originalStatus := storedInst.Status
		// 	storedInst.Status = "exited_or_not_found" // A more descriptive status
		// 	if updateErr := instanceStore.UpdateInstance(instanceID, storedInst); updateErr != nil {
		// 		return fmt.Errorf("instance process not found, and failed to update instance status from '%s' to '%s': %w", originalStatus, storedInst.Status, updateErr)
		// 	}
		// 	fmt.Printf("Instance '%s' status updated to '%s'.\n", instanceID, storedInst.Status)
		// 	return nil // Considered handled, process is not running.
		// }

		// Send a SIGTERM signal (graceful shutdown) to the entire process group
		// Note the negative PID to target the process group.
		if err := syscall.Kill(-storedInst.PID, syscall.SIGTERM); err != nil {
			// If signaling fails, check if it's because the process group doesn't exist (already stopped/gone)
			if errors.Is(err, syscall.ESRCH) { // ESRCH: No such process
				fmt.Printf("%sProcess group with PGID %s%d%s for instance '%s%s%s' not found. It might have already exited.%s\n",
					colorGrayStop, colorBoldStop, storedInst.PID, colorResetStop, colorBoldStop, instanceID, colorResetStop, colorResetStop)
				originalStatus := storedInst.Status
				storedInst.Status = "exited_or_not_found"
				storedInst.StopTime = time.Now().UTC()
				if updateErr := instanceStore.UpdateInstance(instanceID, storedInst); updateErr != nil {
					return fmt.Errorf("process group not found, and failed to update instance status from '%s' to '%s%s%s': %w",
						originalStatus, colorGrayStop, storedInst.Status, colorResetStop, updateErr)
				}
				fmt.Printf("%sInstance '%s%s%s' status updated to '%s%s%s'.%s\n",
					colorGreenStop, colorBoldStop, instanceID, colorResetStop, colorGrayStop, storedInst.Status, colorResetStop, colorResetStop)
				return nil // Considered handled.
			}
			// Other errors (e.g., permission issues)
			return fmt.Errorf("failed to send SIGTERM to process group PGID %s%d%s for instance '%s%s%s': %w. Current status remains '%s%s%s'.",
				colorBoldStop, storedInst.PID, colorResetStop, colorBoldStop, instanceID, colorResetStop, err, colorYellowStop, storedInst.Status, colorResetStop)
		}

		fmt.Printf("%sSent SIGTERM to process group of instance '%s%s%s' (PGID: %s%d%s).%s\n",
			colorGreenStop, colorBoldStop, storedInst.ID, colorResetStop, colorBoldStop, storedInst.PID, colorResetStop, colorResetStop)

		// Update the status in the store
		originalStatus := storedInst.Status
		storedInst.Status = "stopping"
		storedInst.StopTime = time.Now().UTC()
		if err := instanceStore.UpdateInstance(instanceID, storedInst); err != nil {
			return fmt.Errorf("signal sent, but failed to update instance status from '%s' to '%s%s%s': %w",
				originalStatus, colorYellowStop, storedInst.Status, colorResetStop, err)
		}

		fmt.Printf("%sInstance '%s%s%s' status updated to '%s%s%s'. The process is now attempting to shut down gracefully.%s\n",
			colorGreenStop, colorBoldStop, instanceID, colorResetStop, colorYellowStop, storedInst.Status, colorResetStop, colorResetStop)
		fmt.Println("Use 'gitserve list' to check its final status after a short while.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	// Potential future flags:
	// stopCmd.Flags().BoolP("force", "f", false, "Force stop the instance (SIGKILL)")
	// stopCmd.Flags().DurationP("timeout", "t", 0, "Timeout to wait for graceful shutdown before force stopping")
}
