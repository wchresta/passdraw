package runner_test

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/wchresta/passdraw/pkg/runner"
)

func TestRun_SinglePartition(t *testing.T) {
	for _, tc := range []struct {
		name      string
		numUsers  int
		numPasses int
		wantProb  float64
	}{
		{
			name:      "300 users, 300 passes",
			numUsers:  300,
			numPasses: 300,
			wantProb:  1.0,
		},
		{
			name:      "50 users, 30 passes",
			numUsers:  50,
			numPasses: 30,
			wantProb:  30.0 / 50.0,
		},
		{
			name:      "500 users, 300 passes",
			numUsers:  500,
			numPasses: 300,
			wantProb:  300.0 / 500.0,
		},
		{
			name:      "500 users, 30 passes",
			numUsers:  500,
			numPasses: 30,
			wantProb:  30.0 / 500.0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			users := mkFreeUsers("TestPartition", "Free", tc.numUsers)

			rand := rand.New(rand.NewSource(5544332211))
			r := runner.NewWithRand(users, rand)
			allowDelta := 0.02 // Allow 2% difference from expected outcome

			availability := []runner.Availability{{
				Available: tc.numPasses,
				Partition: "TestPartition",
			}}
			for u, prob := range runStats(t, r, availability, 10000) {
				if !strings.HasPrefix(string(u), "Free") {
					continue
				}

				if diff := math.Abs(prob - tc.wantProb); diff > allowDelta {
					t.Errorf("Run produced unexpected probabilities: user=%s got %f, want %f, diff: %f", u, prob, tc.wantProb, diff)
				}
			}
		})
	}
}

func runStats(t *testing.T, r *runner.Runner, availabilities []runner.Availability, runCount int) map[runner.UserID]float64 {
	passes := make(map[runner.UserID]int)
	for i := 0; i < runCount; i++ {
		result, err := r.Run(availabilities)
		if err != nil {
			t.Fatalf("Run failed unexpectedly: %s", err)
		}
		for _, partUsers := range result.Passes {
			for _, userID := range partUsers {
				passes[userID] += 1
			}
		}
	}

	stats := make(map[runner.UserID]float64)
	for u, c := range passes {
		stats[u] = float64(c) / float64(runCount)
	}
	return stats
}

func mkUser(partition string, name string, dependencies ...runner.UserID) *runner.User {
	var deps []runner.UserID
	for _, d := range dependencies {
		deps = append(deps, runner.UserID(d))
	}
	return &runner.User{
		Partition: runner.Partition(partition),
		ID:        runner.UserID(name),
		Deps:      deps,
	}
}

func mkFreeUsers(partition string, prefix string, num int) []*runner.User {
	var users []*runner.User
	for i := 0; i < num; i++ {
		users = append(users, mkUser(partition, fmt.Sprintf("%s%d", prefix, i)))
	}
	return users
}

func mkUsersWithDeps(rand *rand.Rand, partition string, prefix string, num int, deps []runner.UserID, numDeps int) []*runner.User {
	var users []*runner.User
	deps = slices.Clone(deps)
	for i := 0; i < num; i++ {
		rand.Shuffle(len(deps), func(i, j int) { deps[i], deps[j] = deps[j], deps[i] })
		users = append(users, mkUser(partition, fmt.Sprintf("%s%d", prefix, i), deps[:numDeps]...))
	}
	return users
}

func mkUserCouple(partitionLeft string, partitionRight string, name string) []*runner.User {
	return []*runner.User{
		mkUser(partitionLeft, name+"L", runner.UserID(name+"R")),
		mkUser(partitionRight, name+"R", runner.UserID(name+"L")),
	}
}
