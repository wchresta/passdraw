/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"maps"
	"slices"

	"github.com/spf13/cobra"
	"github.com/wchresta/passdraw/pkg/runner"
)

type simulateCmd struct {
	passes int
	runs   int
}

func init() {
	cmd := simulateCmd{}

	var cobraCmd = &cobra.Command{
		Use:   "simulate",
		Short: "Do multiple runs and get statistics",
		Run:   cmd.Simulate,
	}

	rootCmd.AddCommand(cobraCmd)

	cobraCmd.Flags().IntVar(&cmd.passes, "passes", 5, "Amount of passes to hand out")
	cobraCmd.Flags().IntVar(&cmd.runs, "runs", 1000000, "How many runs")
}

func (c *simulateCmd) Simulate(cmd *cobra.Command, args []string) {
	if c.passes <= 0 {
		cmd.PrintErrln("--passes cannot be 0")
		return
	}

	run := runner.New([]*runner.User{
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
		{ID: "D01", Deps: []runner.UserID{"S0", "S1"}},
		{ID: "D12", Deps: []runner.UserID{"S1", "S2"}},
		{ID: "D23", Deps: []runner.UserID{"S2", "S3"}},
	})

	passes := make(map[runner.UserID]int)
	refusals := make(map[runner.UserID]int)
	for i := 0; i <= c.runs; i++ {
		solution, err := run.Run(c.passes)
		if err != nil {
			cmd.PrintErrf("Run failed: %s", err)
			return
		}

		for _, pass := range solution.Passes {
			passes[pass] += 1
		}
		for _, ref := range solution.Refusals {
			refusals[ref] += 1
		}
	}

	passKeys := slices.Collect(maps.Keys(passes))
	slices.Sort(passKeys)

	cmd.Printf("Performed %d runs; here are the statistics:\n", c.runs)
	for _, p := range passKeys {
		n := passes[p]
		cmd.Printf("User %s got a total of %5d passes; probability of %2.1f%%\n", p, n, float64(n)/float64(c.runs)*100)
	}
}
