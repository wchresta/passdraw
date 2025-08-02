/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wchresta/passdraw/pkg/runner"
)

type runCmd struct {
	passes int
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

	cobraCmd.Flags().IntVar(&cmd.passes, "passes", 0, "Amount of passes to hand out")
}

func (c *runCmd) Run(cmd *cobra.Command, args []string) {
	if c.passes <= 0 {
		cmd.PrintErrln("--passes cannot be 0")
		return
	}

	partition := runner.Partition("All")
	r := runner.New([]*runner.User{
		{Partition: partition, ID: "S0"},
		{Partition: partition, ID: "S1"},
		{Partition: partition, ID: "S2"},
		{Partition: partition, ID: "S3"},
		{Partition: partition, ID: "S4"},
		{Partition: partition, ID: "S5"},
		{Partition: partition, ID: "D0", Deps: []runner.UserID{"S0"}},
		{Partition: partition, ID: "D1", Deps: []runner.UserID{"S1"}},
		{Partition: partition, ID: "D2", Deps: []runner.UserID{"S2"}},
		{Partition: partition, ID: "D3", Deps: []runner.UserID{"S3"}},
		{Partition: partition, ID: "D4", Deps: []runner.UserID{"S4"}},
		{Partition: partition, ID: "D5", Deps: []runner.UserID{"S5"}},
	})

	solution, err := r.Run([]runner.Availability{{Partition: partition, Available: c.passes}})
	if err != nil {
		cmd.PrintErrf("Run failed: %s", err)
		return
	}

	cmd.Printf("Executed Run for %d passes, here are the results:\n", c.passes)
	for partName, partPass := range solution.Passes {
		cmd.Printf("Got a total of %d passes for partition %s:\n", len(partPass), partName)
		for _, pass := range partPass {
			cmd.Println(" " + pass)
		}
	}
}
