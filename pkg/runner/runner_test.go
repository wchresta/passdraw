package runner_test

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
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
			partitionPasses := runStats(t, r, availability, 10000)["TestPartition"]
			for u, prob := range partitionPasses {
				if diff := math.Abs(prob - tc.wantProb); diff > allowDelta {
					t.Errorf("Run produced unexpected probabilities: user=%s got %f, want %f, diff: %f", u, prob, tc.wantProb, diff)
				}
			}
		})
	}
}

// We run all above tests at the same time.
func TestRun_IndependentPartitions(t *testing.T) {
	var users []*runner.User
	var availability []runner.Availability
	rand := rand.New(rand.NewSource(5544332211))
	allowDelta := 0.02 // Allow 2% difference from expected outcome

	type partTest struct {
		numUsers  int
		numPasses int
		wantProb  float64
	}
	partitionTests := make(map[runner.Partition]partTest)

	partitionTests["300,300"] = partTest{
		numUsers:  300,
		numPasses: 300,
		wantProb:  1.0,
	}

	partitionTests["50,30"] = partTest{
		numUsers:  50,
		numPasses: 30,
		wantProb:  30.0 / 50.0,
	}

	partitionTests["500,300"] = partTest{
		numUsers:  500,
		numPasses: 300,
		wantProb:  300.0 / 500.0,
	}

	partitionTests["500,30"] = partTest{
		numUsers:  500,
		numPasses: 30,
		wantProb:  30.0 / 500.0,
	}

	for partName, partTest := range partitionTests {
		users = append(users, mkFreeUsers(string(partName), "Free-"+string(partName), partTest.numUsers)...)

		availability = append(availability, []runner.Availability{
			{
				Available: partTest.numPasses,
				Partition: "LeftPartition",
			},
			{
				Available: partTest.numPasses,
				Partition: "RightPartitions",
			},
		}...)
	}

	r := runner.NewWithRand(users, rand)

	for partition, partProbs := range runStats(t, r, availability, 10000) {
		partTest := partitionTests[partition]
		for u, prob := range partProbs {
			if diff := math.Abs(prob - partTest.wantProb); diff > allowDelta {
				t.Errorf("Run produced unexpected probabilities: user=%s got %f, want %f, diff: %f", u, prob, partTest.wantProb, diff)
			}
		}
	}
}

// We run all above tests at the same time.
func TestRun_Weights(t *testing.T) {
	rand := rand.New(rand.NewSource(5544332211))

	// Add free users with weight 1
	var users []*runner.User
	users = append(users, mkFreeUsers("Test", "Free", 8)...)
	// Add one user with weight 2
	users = append(users, &runner.User{
		Partition: "Test",
		ID:        "Weighted",
		Weight:    2.0,
	})

	r := runner.NewWithRand(users, rand)

	availability := []runner.Availability{{Partition: "Test", Available: 3}}

	allowDelta := 0.02 // Allow 2% difference from expected outcome
	for _, partProbs := range runStats(t, r, availability, 10000) {
		for uid, prob := range partProbs {
			want := 3. / 10
			if uid == "Weighted" {
				want = 6. / 10
			}

			if diff := math.Abs(prob - want); diff > allowDelta {
				t.Errorf("Run produced unexpected probabilities: user=%s got %f, want %f, diff: %f", uid, prob, want, diff)
			}
		}
	}
}

func runStats(t *testing.T, r *runner.Runner, availabilities []runner.Availability, runCount int) map[runner.Partition]map[runner.UserID]float64 {
	passes := make(map[runner.Partition]map[runner.UserID]int)
	for i := 0; i < runCount; i++ {
		result, err := r.Run(availabilities)
		if err != nil {
			t.Fatalf("Run failed unexpectedly: %s", err)
		}
		for partition, partUsers := range result.Passes {
			if _, ok := passes[partition]; !ok {
				passes[partition] = make(map[runner.UserID]int)
			}
			for _, userID := range partUsers {
				passes[partition][userID] += 1
			}
		}
	}

	stats := make(map[runner.Partition]map[runner.UserID]float64)
	for partition, partStats := range passes {
		stats[partition] = make(map[runner.UserID]float64)
		for u, c := range partStats {
			stats[partition][u] = float64(c) / float64(runCount)
		}
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
