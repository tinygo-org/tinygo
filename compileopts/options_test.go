package compileopts_test

import (
	"errors"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestOptions_Verify(t *testing.T) {
	testCases := []struct {
		name          string
		opts          compileopts.Options
		expectedError error
	}{
		{
			name: "it returns no error if Options is empty",
			opts: compileopts.Options{},
		},
		{
			name: "it returns an error if gc option is not valid",
			opts: compileopts.Options{
				GC: "incorrect",
			},
			expectedError: compileopts.ErrGCInvalidOption,
		},
		{
			name: "it returns no error if gc option is 'none'",
			opts: compileopts.Options{
				GC: "none",
			},
		},
		{
			name: "it returns no error if gc option is 'leaking'",
			opts: compileopts.Options{
				GC: "leaking",
			},
		},
		{
			name: "it returns no error if gc option is 'extalloc'",
			opts: compileopts.Options{
				GC: "extalloc",
			},
		},
		{
			name: "it returns no error if gc option is 'conservative'",
			opts: compileopts.Options{
				GC: "conservative",
			},
		},
		{
			name: "it returns no error if gc option is empty",
			opts: compileopts.Options{
				GC: "",
			},
		},
		{
			name: "it returns an error if scheduler option is not valid",
			opts: compileopts.Options{
				Scheduler: "incorrect",
			},
			expectedError: compileopts.ErrSchedulerInvalidOption,
		},
		{
			name: "it returns no error if scheduler option is 'tasks'",
			opts: compileopts.Options{
				Scheduler: "task",
			},
		},
		{
			name: "it returns no error if scheduler option is 'coroutines'",
			opts: compileopts.Options{
				Scheduler: "coroutines",
			},
		},
		{
			name: "it returns no error if scheduler option is empty",
			opts: compileopts.Options{
				Scheduler: "",
			},
		},
		{
			name: "it returns an error if printSize option is not valid",
			opts: compileopts.Options{
				PrintSizes: "invalid",
			},
			expectedError: compileopts.ErrPrintSizeInvalidOption,
		},
		{
			name: "it returns no error if printSize option is 'none'",
			opts: compileopts.Options{
				PrintSizes: "none",
			},
		},
		{
			name: "it returns no error if printSize option is 'short'",
			opts: compileopts.Options{
				PrintSizes: "short",
			},
		},
		{
			name: "it returns no error if printSize option is 'full'",
			opts: compileopts.Options{
				PrintSizes: "full",
			},
		},
		{
			name: "it returns no error if printSize option is empty",
			opts: compileopts.Options{
				PrintSizes: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Verify()
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expecting %v, got %v", tc.expectedError, err)
			}
		})
	}
}
