// Package input allows configuration of runners via convenient formats.
package input

import (
	"encoding/json"
	"fmt"

	"github.com/wchresta/passdraw/pkg/runner"
)

type User struct {
	ID   runner.UserID
	Deps []runner.UserID `json:",omitempty"`
}

type RunConfig struct {
	Passes map[runner.Partition]int
	Users  map[runner.Partition][]User
}

func NewFromJSON(b []byte) (*RunConfig, error) {
	conf := RunConfig{}
	if err := json.Unmarshal(b, &conf); err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}
	return &conf, nil
}

func (r *RunConfig) validate() error {
	seenPartitions := make(map[runner.Partition]bool)
	for part := range r.Users {
		if _, ok := seenPartitions[part]; ok {
			return fmt.Errorf("format error: partition %s is listed in users multiple times", part)
		}
		seenPartitions[part] = false
	}

	for part := range r.Passes {
		if _, ok := seenPartitions[part]; !ok {
			return fmt.Errorf("value error: found partition %s with passes but no users", part)
		}
		seenPartitions[part] = true
	}

	for part, seen := range seenPartitions {
		if !seen {
			return fmt.Errorf("value error: found partition %s with users but no passes", part)
		}
	}

	return nil
}

func (r *RunConfig) Runner() *runner.Runner {
	var users []*runner.User
	for part, partUsers := range r.Users {
		partition := runner.Partition(part)
		for _, u := range partUsers {
			users = append(users, &runner.User{
				Partition: partition,
				ID:        u.ID,
				Deps:      u.Deps,
			})
		}
	}
	return runner.New(users)
}

func (r *RunConfig) Availabilities() []runner.Availability {
	var availabilities []runner.Availability
	for part, passes := range r.Passes {
		availabilities = append(availabilities, runner.Availability{
			Partition: runner.Partition(part),
			Available: passes,
		})
	}
	return availabilities
}
