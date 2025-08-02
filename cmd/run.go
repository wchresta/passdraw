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

	runner := runner.New([]*runner.User{
		{ID: "S0"},
		{ID: "S1"},
		{ID: "S2"},
		{ID: "S3"},
		{ID: "S4"},
		{ID: "S5"},
		{ID: "D0", Deps: []runner.UserID{"S0"}},
		{ID: "D1", Deps: []runner.UserID{"S1"}},
		{ID: "D2", Deps: []runner.UserID{"S2"}},
		{ID: "D3", Deps: []runner.UserID{"S3"}},
		{ID: "D4", Deps: []runner.UserID{"S4"}},
		{ID: "D5", Deps: []runner.UserID{"S5"}},
	})

	solution, err := runner.Run(c.passes)
	if err != nil {
		cmd.PrintErrf("Run failed: %s", err)
		return
	}

	cmd.Printf("Executed Run for %d passes, here are the results:\n", c.passes)
	cmd.Printf("Got a total of %d passes for:\n", len(solution.Passes))
	for _, pass := range solution.Passes {
		cmd.Println(" " + pass)
	}
	cmd.Printf("With the following %d refusals:\n", len(solution.Refusals))
	for _, pass := range solution.Refusals {
		cmd.Println(" " + pass)
	}
}
