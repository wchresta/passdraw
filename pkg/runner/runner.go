package runner

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strings"

	"github.com/wchresta/passdraw/pkg/log"
)

type UserID string

type Partition string

type User struct {
	ID        UserID
	Partition Partition
	Deps      []UserID

	// Weight can change how likely it is for a user to get a pass.
	// A number below 1 reduces changes to get a pass, number above 1 increase them.
	// If 0; defaults to 1.
	// Set to a number below 0 to guarantee a refusal.
	Weight float64
}

type Availability struct {
	Partition Partition
	Available int
}

type Runner struct {
	// Managed by New
	userByID map[UserID]User
	rand     *rand.Rand

	// Managed by reset/Run
	usersInPartition map[Partition][]UserID
	dependees        map[UserID][]UserID
	candidates       map[Partition]map[UserID]bool
	candidateWeights map[Partition]float64
}

type Solution struct {
	Passes map[Partition][]UserID
}

func New(users []User) *Runner {
	return NewWithRand(users, rand.New(rand.NewSource(rand.Int63())))
}

func NewWithRand(users []User, rand *rand.Rand) *Runner {
	if rand == nil {
		panic("rand cannot be nil")
	}
	userMap := make(map[UserID]User)
	for _, u := range users {
		// The interface exposes 2 = twice as probable to get a pass.
		// However, internally, we work with refusals.
		// So we need to set the weight to 1/2.
		if u.Weight <= 0 {
			u.Weight = 1
		} else {
			u.Weight = 1 / u.Weight
		}
		userMap[u.ID] = u
	}

	return &Runner{
		userByID: userMap,
		rand:     rand,
	}
}

// reset resets the state to before any calculations.
// This makes testing a lot easier.
func (r *Runner) reset() {
	r.usersInPartition = make(map[Partition][]UserID)
	r.candidateWeights = make(map[Partition]float64)
	r.dependees = make(map[UserID][]UserID)
	r.candidates = make(map[Partition]map[UserID]bool)

	for _, u := range r.userByID {
		r.usersInPartition[u.Partition] = append(r.usersInPartition[u.Partition], u.ID)
		r.candidateWeights[u.Partition] += u.Weight

		for _, dep := range u.Deps {
			r.dependees[dep] = append(r.dependees[dep], u.ID)
		}

		if _, ok := r.candidates[u.Partition]; !ok {
			r.candidates[u.Partition] = make(map[UserID]bool)
		}
		r.candidates[u.Partition][u.ID] = true
	}
}

func (r *Runner) Users(partition Partition) []UserID {
	return slices.Clone(r.usersInPartition[partition])
}

func (r *Runner) User(id UserID) User {
	return r.userByID[id]
}

func (r *Runner) IsRefused(id UserID) bool {
	u := r.User(id)
	cand, ok := r.candidates[u.Partition][id]
	if ok {
		return !cand
	}
	return true
}

// Mark user as refused without propagating the refusal.
func (r *Runner) shallowRefuse(id UserID) {
	u := r.User(id)
	delete(r.candidates[u.Partition], u.ID)
	r.candidateWeights[u.Partition] -= u.Weight
}

// refused refuses the user with the given id, and all users that depend on it.
// Returns true if any user was newly refused.
func (r *Runner) refuse(id UserID) bool {
	if r.IsRefused(id) {
		// Already refused
		return false
	}

	r.shallowRefuse(id)
	for _, d := range r.dependees[id] {
		r.refuse(d)
	}
	return true
}

func (r *Runner) Run(availabilities []Availability) (*Solution, error) {
	r.reset()

	availabilitiesByPartition := make(map[Partition]Availability)
	for _, a := range availabilities {
		availabilitiesByPartition[a.Partition] = a
	}

	partitionNeedsRefusals := make(map[Partition]bool)
	for partName := range r.usersInPartition {
		partitionNeedsRefusals[partName] = true
	}

	madeProgress := true
	for madeProgress {
		madeProgress = false

	PartitionLoop:
		for partName, isOpen := range partitionNeedsRefusals {
			if !isOpen {
				continue
			}

			av := availabilitiesByPartition[partName]
			partUsers := r.usersInPartition[partName]

			// Check if this partition is still open.
			// We need to check here, because other partitions might
			// have refused enough users here to close it.
			// Users that are not refused get a pass.
			if av.Available >= len(r.candidates[partName]) {
				// We refused enough users
				partitionNeedsRefusals[partName] = false
				continue
			}

			// Find next refusal
			refusalVal := r.rand.Float64() * r.candidateWeights[partName]
			// Find the refused user
			localWeightSum := 0.0
			for u := range r.candidates[partName] {
				localWeightSum += r.User(u).Weight
				if localWeightSum < refusalVal {
					continue
				}

				// u is the user to be refused!
				// Swap the current user to the end of the pool and then shrink the pool.
				if r.refuse(u) {
					// Only if the current user is not already refused we continue.
					// Other
					madeProgress = true
					continue PartitionLoop
				}
			}

			// Did not break; so we run out of users to refuse.
			log.Warningf("Refused all %d possible users for partition %s\n", len(partUsers), partName)
			partitionNeedsRefusals[partName] = false
		}
	}

	passes := make(map[Partition][]UserID)
	for partName, cand := range r.candidates {
		passes[partName] = slices.Sorted(maps.Keys(cand))
	}
	return &Solution{
		Passes: passes,
	}, nil
}

func UserFromString(s string) (*User, error) {
	userDeps := strings.Split(s, ":")
	user := User{
		ID: UserID(userDeps[0]),
	}
	if len(userDeps) > 1 {
		for dep := range strings.SplitSeq(userDeps[1], ",") {
			if dep == "" {
				return nil, errors.New("user dependency entry cannot be the empty string")
			}
			user.Deps = append(user.Deps, UserID(dep))
		}
	}
	if user.ID == "" {
		return nil, errors.New("user id cannot be empty")
	}
	return &user, nil
}

func UsersFromStrings(lines []string) ([]*User, error) {
	var users []*User
	for i, s := range lines {
		u, err := UserFromString(s)
		if err != nil {
			return nil, fmt.Errorf("parse error on line %d: %w", i, err)
		}
		users = append(users, u)
	}
	return users, nil
}
