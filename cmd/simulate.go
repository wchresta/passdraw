/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"cmp"
	"fmt"
	"iter"
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

	cobraCmd.Flags().IntVar(&cmd.passes, "passes", 10, "Amount of passes to hand out")
	cobraCmd.Flags().IntVar(&cmd.runs, "runs", 1000000, "How many runs")
}

func (c *simulateCmd) Simulate(cmd *cobra.Command, args []string) {
	if c.passes <= 0 {
		cmd.PrintErrln("--passes cannot be 0")
		return
	}

	leaderP := runner.Partition("Leader")
	followP := runner.Partition("Follow")
	users := []*runner.User{
		{Partition: leaderP, ID: "L1"},
		{Partition: leaderP, ID: "L2"},
		{Partition: leaderP, ID: "L3"},
		{Partition: leaderP, ID: "L4"},
		{Partition: leaderP, ID: "L5"},
		{Partition: followP, ID: "F1"},
		{Partition: followP, ID: "F2"},
		{Partition: followP, ID: "F3"},
		{Partition: followP, ID: "F4"},
		{Partition: followP, ID: "F5"},
		{Partition: leaderP, ID: "La->F1", Deps: []runner.UserID{"F1"}},
		{Partition: leaderP, ID: "Lb->F2", Deps: []runner.UserID{"F2"}},
		{Partition: leaderP, ID: "Lc->F3", Deps: []runner.UserID{"F3"}},
		{Partition: followP, ID: "Fa->L3", Deps: []runner.UserID{"L3"}},
		{Partition: followP, ID: "Fb->L4", Deps: []runner.UserID{"L4"}},
		{Partition: followP, ID: "Fc->L5", Deps: []runner.UserID{"L5"}},
		{Partition: leaderP, ID: "Lx->L1", Deps: []runner.UserID{"L1"}},
		{Partition: leaderP, ID: "Ly->L2", Deps: []runner.UserID{"L2"}},
		{Partition: followP, ID: "Fx->F2", Deps: []runner.UserID{"F2"}},
		{Partition: followP, ID: "Fy->F3", Deps: []runner.UserID{"F3"}},
		{Partition: leaderP, ID: "Lp->Fx", Deps: []runner.UserID{"Fx->F2"}},
		{Partition: leaderP, ID: "Lq->Fy,F3", Deps: []runner.UserID{"Fy->F3", "F3"}},
		{Partition: leaderP, ID: "LC1", Deps: []runner.UserID{"FC1"}},
		{Partition: followP, ID: "FC1", Deps: []runner.UserID{"LC1"}},
		{Partition: leaderP, ID: "LC2", Deps: []runner.UserID{"FC2"}},
		{Partition: followP, ID: "FC2", Deps: []runner.UserID{"LC2"}},
	}
	userCountByPart := make(map[runner.Partition]int)
	for _, u := range users {
		userCountByPart[u.Partition]++
	}
	r := runner.New(users)

	availMap := make(map[runner.Partition]runner.Availability)
	availMap[leaderP] = runner.Availability{Partition: leaderP, Available: c.passes / 2}
	availMap[followP] = runner.Availability{Partition: followP, Available: c.passes - (c.passes / 2)}

	passes := make(map[runner.Partition]map[runner.UserID]int)
	for i := 0; i <= c.runs; i++ {
		solution, err := r.Run(slices.Collect(maps.Values(availMap)))
		if err != nil {
			cmd.PrintErrf("Run failed: %s", err)
			return
		}

		for part, partPass := range solution.Passes {
			if _, ok := passes[part]; !ok {
				passes[part] = make(map[runner.UserID]int)
			}
			for _, pass := range partPass {
				passes[part][pass] += 1
			}
		}
	}

	cmd.Printf("Performed %d runs; here are the statistics:\n", c.runs)
	for part, passes := range sortedKeys(passes) {
		a := availMap[part]
		fmt.Printf("Handed out %d passes to %d users in partition %s\n", a.Available, userCountByPart[part], part)
		totalPasses := 0
		for u, n := range sortedKeys(passes) {
			totalPasses += n
			cmd.Printf("User %-10s got a total of %6d passes; probability of %4.1f%%\n", u, n, float64(n)/float64(c.runs)*100)
		}
		fmt.Printf("Handed out a total of %d passes for partition %s\n", totalPasses, part)
	}
}

func sortedKeys[Map ~map[K]V, K cmp.Ordered, V any](m Map) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}
