/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wchresta/passdraw/pkg/runner"
)

type runCmd struct {
	availStrings []string
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
}

func (c *runCmd) Run(cmd *cobra.Command, args []string) {
	availMap := make(map[runner.Partition]runner.Availability)
	for _, availStr := range c.availStrings {
		part, passStr, found := strings.Cut(availStr, ":")
		if !found {
			cmd.PrintErrln("--passes must have a `:`. Format `partition:passes`, e.g. `leaders:33`")
			return
		}
		passes, err := strconv.Atoi(passStr)
		if err != nil {
			cmd.PrintErrln("--passes must contain a valid number. Format `partition:passes`, e.g. `leaders:33`")
			return
		}
		availMap[runner.Partition(part)] = runner.Availability{
			Partition: runner.Partition(part),
			Available: passes,
		}
	}

	leaderP := runner.Partition("leader")
	followP := runner.Partition("follow")
	r := runner.New([]*runner.User{
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
	})

	solution, err := r.Run(slices.Collect(maps.Values(availMap)))
	if err != nil {
		cmd.PrintErrf("Run failed: %s", err)
		return
	}

	cmd.Println("Executed Run for the following availabilities:")
	for partName, partPass := range solution.Passes {
		a := availMap[partName]
		cmd.Printf("%s - Handed out %d out of %d passes for partition:\n", partName, len(partPass), a.Available)
		slices.Sort(partPass)
		for _, pass := range partPass {
			cmd.Println(" " + pass)
		}
	}
}
