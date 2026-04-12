package store

import (
	"errors"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func TestIsDuplicateConstraintError(testCase *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique constraint",
			err: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintUnique,
			},
			want: true,
		},
		{
			name: "primary key constraint",
			err: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintPrimaryKey,
			},
			want: true,
		},
		{
			name: "other constraint",
			err: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintForeignKey,
			},
			want: false,
		},
		{
			name: "non sqlite error",
			err:  errors.New("UNIQUE constraint failed: people.email"),
			want: false,
		},
	}

	for _, test := range testCases {
		testCase.Run(test.name, func(testCase *testing.T) {
			if got := isDuplicateConstraintError(test.err); got != test.want {
				testCase.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}
