package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wchresta/passdraw/pkg/input"
	"github.com/wchresta/passdraw/pkg/runner"
)

var (
	leaderFullPasses    int
	leaderPartyPasses   int
	followerFullPasses  int
	followerPartyPasses int

	LFP runner.Partition = "leader_full"
	LPP runner.Partition = "leader_part"
	FFP runner.Partition = "follow_full"
	FPP runner.Partition = "follow_part"

	overbookRatio float64
	partyCouples  int
	fullCouples   int
)

// genCmd represents the base command when called without any subcommands
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "gen generates test input",
	Run:   Run,
}

func main() {
	genCmd.Execute()
}

func init() {
	genCmd.Flags().IntVar(&followerFullPasses, "follower_full_passes", 170, "")
	genCmd.Flags().IntVar(&followerPartyPasses, "follower_party_passes", 85, "")
	genCmd.Flags().IntVar(&leaderFullPasses, "leader_full_passes", 150, "")
	genCmd.Flags().IntVar(&leaderPartyPasses, "leader_party_passes", 85, "")

	genCmd.Flags().Float64Var(&overbookRatio, "overbook_ratio", 1.5, "")
	genCmd.Flags().IntVar(&fullCouples, "full_couples", 30, "")
	genCmd.Flags().IntVar(&partyCouples, "party_couples", 10, "")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := genCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Run(cmd *cobra.Command, args []string) {
	c := input.RunConfig{
		Passes: map[runner.Partition]int{
			LFP: leaderFullPasses,
			LPP: leaderPartyPasses,
			FFP: followerFullPasses,
			FPP: followerPartyPasses,
		},
		Users: make(map[runner.Partition][]input.User),
	}

	numCouples := map[runner.Partition]int{
		LFP: fullCouples,
		FFP: fullCouples,
		LPP: partyCouples,
		FPP: partyCouples,
	}
	couplePartition := map[runner.Partition]runner.Partition{
		LFP: FFP,
		FFP: LFP,
		LPP: FPP,
		FPP: LPP,
	}

	for part, numPass := range c.Passes {
		numUsers := int(float64(numPass) * overbookRatio)
		numCouples := numCouples[part]
		numSolo := numUsers - 2*numCouples

		var users []input.User
		for n := 1; n <= numSolo; n++ {
			userID := fmt.Sprintf("%s-%03d", part, n)
			users = append(users, input.User{
				ID: runner.UserID(userID),
			})
		}

		for n := 1; n <= numCouples; n++ {
			userID := fmt.Sprintf("%s-couple-%02d", part, n)
			partnerID := fmt.Sprintf("%s-couple-%02d", couplePartition[part], n)
			users = append(users, input.User{
				ID:   runner.UserID(userID),
				Deps: []runner.UserID{runner.UserID(partnerID)},
			})
		}

		c.Users[part] = users
	}

	b, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	cmd.OutOrStdout().Write(b)
}
