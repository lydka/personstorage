package store

import (
	"errors"
	"testing"
)

func TestIsDuplicateEmailUniqueConstraintError(testCase *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "email unique constraint",
			err:  errors.New("UNIQUE constraint failed: people.email"),
			want: true,
		},
		{
			name: "external id unique constraint",
			err:  errors.New("UNIQUE constraint failed: people.external_id"),
			want: false,
		},
		{
			name: "non sqlite error",
			err:  errors.New("something else failed"),
			want: false,
		},
	}

	for _, test := range testCases {
		testCase.Run(test.name, func(testCase *testing.T) {
			if got := isDuplicateEmailUniqueConstraintError(test.err); got != test.want {
				testCase.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}
