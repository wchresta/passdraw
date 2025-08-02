package runner

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strings"
)

type UserID string

type User struct {
	ID   UserID
	Deps []UserID
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

type Runner struct {
	userMap   map[UserID]*User
	dependees map[UserID][]UserID
	rand      *rand.Rand

	refusals map[UserID]bool
}

type Solution struct {
	Passes   []UserID
	Refusals []UserID
}

func New(users []*User) *Runner {
	userMap := make(map[UserID]*User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	return &Runner{
		userMap: userMap,
		rand:    rand.New(rand.NewSource(rand.Int63())),
	}
}

func (r *Runner) prepare() {
	r.refusals = make(map[UserID]bool)
	r.dependees = make(map[UserID][]UserID)

	for _, user := range r.userMap {
		for _, dep := range user.Deps {
			r.dependees[dep] = append(r.dependees[dep], user.ID)
		}
	}
}

func (r *Runner) refuse(id UserID) int {
	if _, ok := r.refusals[id]; ok {
		// Already refused
		return 0
	}

	r.refusals[id] = true
	newRefusals := 1
	for _, d := range r.dependees[id] {
		newRefusals += r.refuse(d)
	}
	return newRefusals
}

func (r *Runner) Run(passCount int) (*Solution, error) {
	userIDs := slices.Collect(maps.Keys(r.userMap))
	if passCount >= len(r.userMap) {
		slices.Sort(userIDs)
		return &Solution{
			Passes:   userIDs,
			Refusals: nil,
		}, nil
	}
	r.prepare()

	r.rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})

	for _, uID := range userIDs {
		r.refuse(uID)

		if len(r.refusals)+passCount >= len(userIDs) {
			break
		}
	}

	refusals := slices.Collect(maps.Keys(r.refusals))
	slices.Sort(refusals)

	var passes []UserID
	for _, u := range userIDs {
		if _, ok := r.refusals[u]; !ok {
			passes = append(passes, u)
		}
	}
	slices.Sort(passes)

	return &Solution{
		Passes:   passes,
		Refusals: refusals,
	}, nil
}
