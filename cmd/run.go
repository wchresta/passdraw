/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wchresta/passdraw/pkg/input"
	"github.com/wchresta/passdraw/pkg/runner"
)

type runCmd struct {
	availStrings []string
	inputPath    string
}

func init() {
	cmd := runCmd{}

	var cobraCmd = &cobra.Command{
		Use:   "run",
		Short: "Assign passes to users.",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: cmd.Run,
	}

	rootCmd.AddCommand(cobraCmd)

	cobraCmd.Flags().StringSliceVar(&cmd.availStrings, "passes", nil, "Specify availability of passes for partition; format `partition:passes` e.g. `leaders:33`")
	cobraCmd.Flags().StringVar(&cmd.inputPath, "input", "", "Path to take the input from. Must be in json format")
}

func (c *runCmd) Run(cmd *cobra.Command, args []string) {
	var run *runner.Runner
	var avail []runner.Availability

	availMap, err := availMapFromAvailStrings(c.availStrings)
	if err != nil {
		cmd.PrintErr(err)
		return
	}
	avail = slices.Collect(maps.Values(availMap))

	if c.inputPath != "" {
		inputBs, err := os.ReadFile(c.inputPath)
		if err != nil {
			cmd.PrintErrf("cannot read input file %s: %s", c.inputPath, err)
			return
		}

		conf, err := input.NewFromJSON(inputBs)
		if err != nil {
			cmd.PrintErrf("cannot parse input file %s: %s", c.inputPath, err)
			return
		}

		run = conf.Runner()
		if len(avail) == 0 {
			avail = conf.Availabilities()
			for _, a := range avail {
				availMap[a.Partition] = a
			}
		}
	}

	solution, err := run.Run(avail)
	if err != nil {
		cmd.PrintErrf("Run failed: %s", err)
		return
	}

	cmd.Println("Executed Run for the following availabilities:")
	for partName, partPass := range solution.Passes {
		a := availMap[partName]

		cmd.Printf("%s - Handed out %d out of %d passes for partition:\n", partName, len(partPass), a.Available)
		slices.Sort(partPass)

		hasPass := make(map[runner.UserID]bool)
		for _, u := range run.Users(partName) {
			hasPass[u] = false
		}
		for _, pass := range partPass {
			delete(hasPass, pass)
			cmd.Println(" O " + pass)
		}

		cmd.Printf("%s - The following %d users did not get a pass:\n", partName, len(hasPass))
		for u := range sortedKeys(hasPass) {
			cmd.Println(" x " + u)
		}
	}
}

func availMapFromAvailStrings(availStrings []string) (map[runner.Partition]runner.Availability, error) {
	availMap := make(map[runner.Partition]runner.Availability)
	for _, availStr := range availStrings {
		part, passStr, found := strings.Cut(availStr, ":")
		if !found {
			return nil, fmt.Errorf("--passes must have a `:`. Format `partition:passes`, e.g. `leaders:33`")
		}
		passes, err := strconv.Atoi(passStr)
		if err != nil {
			return nil, fmt.Errorf("--passes must contain a valid number. Format `partition:passes`, e.g. `leaders:33`")
		}
		availMap[runner.Partition(part)] = runner.Availability{
			Partition: runner.Partition(part),
			Available: passes,
		}
	}
	return availMap, nil
}
