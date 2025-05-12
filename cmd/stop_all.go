package cmd

import (
	"errors"
	"fmt"
	"gitserve/internal/storage"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	colorResetStopAll  = "\033[0m"
	colorRedStopAll    = "\033[31m"
	colorGreenStopAll  = "\033[32m"
	colorYellowStopAll = "\033[33m"
	colorGrayStopAll   = "\033[90m"
	colorBoldStopAll   = "\033[1m"
	colorCyanStopAll   = "\033[36m"
)

var (
	stopAllProjectName string
)

var stopAllCmd = &cobra.Command{
	Use:   "stop-all",
	Short: "Stop all running gitserve instances, optionally filtered by project name",
	Long:  `Stops all gitserve instances that are currently in a 'running' state. Can be filtered by project name.`,
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

		if len(instances) == 0 {
			fmt.Printf("%sNo instances found to stop.%s\n", colorYellowStopAll, colorResetStopAll)
			return nil
		}

		var wg sync.WaitGroup
		stoppedCount := 0
		failedToStopCount := 0
		skippedCount := 0

		fmt.Printf("Attempting to stop instances (%sfilter%s: '%s%s%s')...\n",
			colorCyanStopAll, colorResetStopAll,
			colorBoldStopAll, func() string {
				if stopAllProjectName == "" {
					return "none"
				}
				return stopAllProjectName
			}(), colorResetStopAll)

		// Quick fix for counters: Use a channel to collect results.
		type result struct {
			id            string
			name          string
			success       bool
			finalStatus   string
			isSkipped     bool
			skippedReason string
			errorMsg      string
		}

		resultsChan := make(chan result, len(instances))
		activeAttempts := 0

		for _, inst := range instances {
			instanceCopy := inst // Work with a copy for the goroutine
			if stopAllProjectName != "" {
				if instanceCopy.Path == "" {
					resultsChan <- result{id: instanceCopy.ID, name: instanceCopy.Name, isSkipped: true, skippedReason: "missing path for project filtering"}
					continue
				}
				instanceProjectName := filepath.Base(instanceCopy.Path)
				if !strings.EqualFold(instanceProjectName, stopAllProjectName) {
					resultsChan <- result{id: instanceCopy.ID, name: instanceCopy.Name, isSkipped: true, skippedReason: "project name mismatch"}
					continue
				}
			}

			if instanceCopy.Status != "running" {
				resultsChan <- result{id: instanceCopy.ID, name: instanceCopy.Name, isSkipped: true, skippedReason: fmt.Sprintf("status is '%s%s%s', not '%srunning%s'", colorYellowStopAll, instanceCopy.Status, colorResetStopAll, colorGreenStopAll, colorResetStopAll)}
				continue
			}

			if instanceCopy.PID == 0 {
				resultsChan <- result{id: instanceCopy.ID, name: instanceCopy.Name, isSkipped: true, skippedReason: "PID is 0"}
				continue
			}

			activeAttempts++
			wg.Add(1)
			go func(instanceToStop storage.Instance) {
				defer wg.Done()
				cmd.Printf("  Stopping instance %s%s%s (%s) - PGID: %s%d%s...\n",
					colorBoldStopAll, instanceToStop.ID, colorResetStopAll, instanceToStop.Name, colorBoldStopAll, instanceToStop.PID, colorResetStopAll)

				finalStatus := instanceToStop.Status
				currentTime := time.Now().UTC()
				signalErr := syscall.Kill(-instanceToStop.PID, syscall.SIGTERM)

				if signalErr != nil {
					if errors.Is(signalErr, syscall.ESRCH) {
						cmd.Printf("    %sProcess group for %s%s%s (PGID: %s%d%s) not found. Already exited.%s\n",
							colorGrayStopAll, colorBoldStopAll, instanceToStop.ID, colorResetStopAll, colorBoldStopAll, instanceToStop.PID, colorResetStopAll, colorResetStopAll)
						finalStatus = "exited_or_not_found"
						instanceToStop.StopTime = currentTime
					} else {
						resultsChan <- result{id: instanceToStop.ID, name: instanceToStop.Name, success: false, errorMsg: fmt.Sprintf("failed to send SIGTERM to PGID %d: %v", instanceToStop.PID, signalErr)}
						return
					}
				} else {
					cmd.Printf("    %sSent SIGTERM to PGID %s%d%s for %s%s%s.%s\n",
						colorGreenStopAll, colorBoldStopAll, instanceToStop.PID, colorResetStopAll, colorBoldStopAll, instanceToStop.ID, colorResetStopAll, colorResetStopAll)
					finalStatus = "stopping"
					instanceToStop.StopTime = currentTime
				}

				instanceToStop.Status = finalStatus
				if updateErr := instanceStore.UpdateInstance(instanceToStop.ID, instanceToStop); updateErr != nil {
					resultsChan <- result{id: instanceToStop.ID, name: instanceToStop.Name, success: false, finalStatus: finalStatus, errorMsg: fmt.Sprintf("failed to update store to '%s': %v", finalStatus, updateErr)}
				} else {
					resultsChan <- result{id: instanceToStop.ID, name: instanceToStop.Name, success: true, finalStatus: finalStatus}
				}
			}(instanceCopy)
		}

		wg.Wait()
		close(resultsChan)

		// Process results from channel
		for res := range resultsChan {
			if res.isSkipped {
				cmd.Printf("  %sSkipped %s%s%s (%s): %s%s\n", colorGrayStopAll, colorBoldStopAll, res.id, colorResetStopAll, res.name, res.skippedReason, colorResetStopAll)
				skippedCount++
			} else if res.success {
				color := colorGreenStopAll
				if res.finalStatus == "stopping" {
					color = colorYellowStopAll
				}
				cmd.Printf("  %sInstance %s%s%s (%s) processed. Final status: %s%s%s%s\n",
					colorGreenStopAll, colorBoldStopAll, res.id, colorResetStopAll, res.name, color, res.finalStatus, colorResetStopAll, colorResetStopAll)
				if res.finalStatus == "stopping" || res.finalStatus == "exited_or_not_found" {
					stoppedCount++
				}
			} else {
				cmd.PrintErrf("  %sError processing %s%s%s (%s): %s. Status attempted: '%s%s%s'.%s\n",
					colorRedStopAll, colorBoldStopAll, res.id, colorResetStopAll, res.name, res.errorMsg, colorYellowStopAll, res.finalStatus, colorResetStopAll, colorResetStopAll)
				failedToStopCount++
			}
		}

		fmt.Printf("\n%s--- Stop All Summary ---%s\n", colorBoldStopAll, colorResetStopAll)
		fmt.Printf("  %sSuccessfully signaled/processed: %s%d%s%s\n", colorGreenStopAll, colorBoldStopAll, stoppedCount, colorResetStopAll, colorResetStopAll)
		fmt.Printf("  %sFailed to stop/update:         %s%d%s%s\n", colorRedStopAll, colorBoldStopAll, failedToStopCount, colorResetStopAll, colorResetStopAll)
		fmt.Printf("  %sSkipped:                       %s%d%s%s\n", colorGrayStopAll, colorBoldStopAll, skippedCount, colorResetStopAll, colorResetStopAll)
		fmt.Println("Use 'gitserve list' to verify final statuses and for pruning.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopAllCmd)
	stopAllCmd.Flags().StringVarP(&stopAllProjectName, "project", "p", "", "Filter instances by project name (derived from the repository directory name)")
}
