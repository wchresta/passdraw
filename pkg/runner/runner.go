package runner

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"
)

type UserID string

type Partition string

type User struct {
	ID        UserID
	Partition Partition
	Deps      []UserID
}

type Availability struct {
	Partition Partition
	Available int
}

type Runner struct {
	userByID         map[UserID]*User
	usersByPartition map[Partition][]UserID
	dependees        map[UserID][]UserID
	rand             *rand.Rand

	refusals map[Partition]map[UserID]bool
}

type Solution struct {
	Passes map[Partition][]UserID
}

func New(users []*User) *Runner {
	return NewWithRand(users, rand.New(rand.NewSource(rand.Int63())))
}

func NewWithRand(users []*User, rand *rand.Rand) *Runner {
	if rand == nil {
		panic("rand cannot be nil")
	}
	passUsers := make(map[Partition][]UserID)
	userMap := make(map[UserID]*User)
	for _, u := range users {
		userMap[u.ID] = u
		passUsers[u.Partition] = append(passUsers[u.Partition], u.ID)
	}
	return &Runner{
		userByID:         userMap,
		usersByPartition: passUsers,
		rand:             rand,
	}
}

func (r *Runner) User(id UserID) *User {
	return r.userByID[id]
}

func (r *Runner) IsRefused(id UserID) bool {
	u := r.User(id)
	refusals, ok := r.refusals[u.Partition]
	if !ok {
		return false
	}
	_, refused := refusals[id]
	return refused
}

// Mark user as refused without propagating the refusal.
func (r *Runner) shallowRefuse(id UserID) {
	u := r.User(id)
	_, ok := r.refusals[u.Partition]
	if !ok {
		r.refusals[u.Partition] = make(map[UserID]bool)
	}
	r.refusals[u.Partition][id] = true
}

func (r *Runner) prepare() {
	r.refusals = make(map[Partition]map[UserID]bool)
	r.dependees = make(map[UserID][]UserID)

	for _, user := range r.userByID {
		for _, dep := range user.Deps {
			r.dependees[dep] = append(r.dependees[dep], user.ID)
		}
	}
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
	r.prepare()

	availabilitiesByPartition := make(map[Partition]Availability)
	for _, a := range availabilities {
		availabilitiesByPartition[a.Partition] = a
	}

	// We shuffle the users for every pass type
	// The users will be in the order of their refusal, meaning users
	// in the beginning will be refused, while users toward the end will not.
	partitionNeedsRefusals := make(map[Partition]bool)
	for partName, partUsers := range r.usersByPartition {
		partitionNeedsRefusals[partName] = true
		r.rand.Shuffle(len(partUsers), func(i, j int) {
			partUsers[i], partUsers[j] = partUsers[j], partUsers[i]
		})
	}

	refusalIdx := make(map[Partition]int)
	for partName, _ := range partitionNeedsRefusals {
		refusalIdx[partName] = 0
	}

	madeProgress := true
	for madeProgress {
		madeProgress = false
		for partName, isOpen := range partitionNeedsRefusals {
			if !isOpen {
				continue
			}

			av := availabilitiesByPartition[partName]
			partUsers := r.usersByPartition[partName]

			// Check if this partition is still open.
			// We need to check here, because other partitions might
			// have refused enough users here to close it.
			// Users that are not refused get a pass.
			if av.Available >= len(partUsers)-len(r.refusals[partName]) {
				// We refused enough users
				partitionNeedsRefusals[partName] = false
				continue
			}

			for idx := refusalIdx[partName]; idx < len(partUsers); idx++ {
				if !r.refuse(partUsers[idx]) {
					// Cannot refuse users as they are refused already.
					continue
				}

				// Made new refusals
				madeProgress = true
				break
			}
			// Did not break; so we run out of users to refuse.
			partitionNeedsRefusals[partName] = false
		}
	}

	passes := make(map[Partition][]UserID)
	for uID, user := range r.userByID {
		if !r.IsRefused(uID) {
			passes[user.Partition] = append(passes[user.Partition], uID)
		}
	}
	for _, partPasses := range passes {
		slices.Sort(partPasses)
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
