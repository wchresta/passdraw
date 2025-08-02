package runner_test

import (
	"strings"
	"testing"

	"github.com/wchresta/passdraw/pkg/runner"
)

func TestRun(t *testing.T) {
	usersString := `0
1
2
3
4
5
6
7
8
9
10:1
11:2
12:3
13:4
14:0,1,2,3,4,5,6,7,8,9`

	users, err := runner.UsersFromStrings(strings.Split(usersString, "\n"))
	if err != nil {
		t.Fatalf("UsersFromString: %s", err)
	}

	r := runner.New(users)

	solution, err := r.Run(10)
	if err != nil {
		t.Errorf("Run returned unexpected error: %s", err)
	}

	seen14Refused := false
	for _, ref := range solution.Refusals {
		if ref == "14" {
			seen14Refused = true
		}
	}
	if !seen14Refused {
		t.Errorf("Run did not refuse 14")
	}
}
