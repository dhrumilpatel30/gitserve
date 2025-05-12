package cmd

import (
	"errors"
	"fmt"
	"gitserve/internal/storage"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

const pruneAge = 1 * time.Minute // Instances stopped longer than this will be pruned by list

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances, update status, and prune old stopped instances",
	Long:  `Displays gitserve instances, updates status based on PID liveness, and prunes instances stopped for more than ` + pruneAge.String() + ` along with their workspaces.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		storeDataPath := filepath.Join(homeDir, ".gitserve", "store")

		instanceStore, err := storage.NewJSONInstanceStore(storeDataPath)
		if err != nil {
			return fmt.Errorf("failed to initialize instance store: %w", err)
		}

		instances, err := instanceStore.GetAllInstances()
		if err != nil {
			return fmt.Errorf("failed to retrieve instances: %w", err)
		}

		var instancesToDisplay []storage.Instance
		processedTime := time.Now().UTC()

		for _, inst := range instances {
			currentInst := inst
			needsStoreUpdate := false
			originalStatus := currentInst.Status

			if (strings.ToLower(currentInst.Status) == "running" || strings.ToLower(currentInst.Status) == "stopping") && currentInst.PID > 0 {
				process, _ := os.FindProcess(currentInst.PID) // Error can be ignored here, Signal will fail if PID is bad.
				if err := process.Signal(syscall.Signal(0)); err != nil {
					if errors.Is(err, os.ErrProcessDone) || strings.Contains(strings.ToLower(err.Error()), "no such process") {
						if strings.ToLower(originalStatus) == "running" {
							currentInst.Status = "exited_unexpectedly"
						} else { // Was "stopping"
							currentInst.Status = "stopped"
						}
						currentInst.StopTime = processedTime
						needsStoreUpdate = true
						cmd.Printf("(Auto-updated ID %s: status '%s' -> '%s', PID %d not found)\n", currentInst.ID, originalStatus, currentInst.Status, currentInst.PID)
					}
				}
			}

			// Pruning logic
			isTerminalStatus := false
			switch strings.ToLower(currentInst.Status) {
			case "stopped", "exited_unexpectedly", "failed", "exited_or_not_found", "error_pid_zero":
				isTerminalStatus = true
			}

			if isTerminalStatus {
				if currentInst.StopTime.IsZero() { // If StopTime wasn't set (e.g. old record or manual edit)
					currentInst.StopTime = processedTime // Set it now
					needsStoreUpdate = true
				}
				if time.Since(currentInst.StopTime.In(time.UTC)) > pruneAge {
					cmd.Printf("(Pruning old instance ID %s: status '%s', stopped at %s)...\n", currentInst.ID, currentInst.Status, currentInst.StopTime.Local().Format(time.RFC3339))
					if errDel := instanceStore.DeleteInstance(currentInst.ID); errDel != nil {
						cmd.PrintErrf("  Error deleting instance %s from store: %v\n", currentInst.ID, errDel)
					} else {
						cmd.Printf("  Instance %s removed from store.\n", currentInst.ID)
					}
					if currentInst.Path != "" {
						if errRm := os.RemoveAll(currentInst.Path); errRm != nil {
							cmd.PrintErrf("  Error cleaning up workspace '%s': %v\n", currentInst.Path, errRm)
						} else {
							cmd.Printf("  Workspace '%s' cleaned up.\n", currentInst.Path)
						}
					}
					needsStoreUpdate = false // Already deleted, no further update needed for this one.
					continue                 // Skip adding to display list
				}
			}

			if needsStoreUpdate {
				if updateErr := instanceStore.UpdateInstance(currentInst.ID, currentInst); updateErr != nil {
					cmd.PrintErrf("Error updating store for instance %s: %v\n", currentInst.ID, updateErr)
					// if update fails, add original inst to display
					instancesToDisplay = append(instancesToDisplay, inst)
					continue
				}
			}
			instancesToDisplay = append(instancesToDisplay, currentInst)
		}

		if len(instancesToDisplay) == 0 {
			fmt.Println("No active or recently stopped instances found.")
			return nil
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape) // Pad 2, strip escape for color calcs
		fmt.Fprintln(writer, colorBold+"ID\tNAME\tPID\tPORT\tSTATUS\tPATH\tSTART TIME\tSTOP TIME"+colorReset)
		fmt.Fprintln(writer, colorBold+"--\t----\t---\t----\t------\t----\t----------\t---------"+colorReset)

		for _, instToDisplay := range instancesToDisplay {
			startTimeFormatted := "N/A"
			if !instToDisplay.StartTime.IsZero() {
				startTimeFormatted = instToDisplay.StartTime.Local().Format("01-02 15:04:05")
			}
			stopTimeFormatted := "N/A"
			if !instToDisplay.StopTime.IsZero() {
				stopTimeFormatted = instToDisplay.StopTime.Local().Format("01-02 15:04:05")
			}

			statusColor := colorReset
			switch strings.ToLower(instToDisplay.Status) {
			case "running":
				statusColor = colorGreen
			case "stopping":
				statusColor = colorYellow
			case "stopped", "exited_or_not_found":
				statusColor = colorGray
			case "failed", "error_pid_zero", "exited_unexpectedly":
				statusColor = colorRed
			default:
				statusColor = colorCyan
			}
			coloredStatus := statusColor + instToDisplay.Status + colorReset

			displayPath := instToDisplay.Path
			maxPathLen := 35
			if len(displayPath) > maxPathLen {
				displayPath = "..." + displayPath[len(displayPath)-maxPathLen+3:]
			}

			fmt.Fprintf(writer, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\n",
				instToDisplay.ID,
				instToDisplay.Name,
				instToDisplay.PID,
				instToDisplay.Port,
				coloredStatus,
				displayPath,
				startTimeFormatted,
				stopTimeFormatted,
			)
		}
		writer.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
