package compileopts_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestVerifyOptions(t *testing.T) {
	testCases := []struct {
		name          string
		opts          compileopts.Options
		expectedError error
	}{
		{
			name: "OptionsEmpty",
			opts: compileopts.Options{},
		},
		{
			name: "InvalidGCOption",
			opts: compileopts.Options{
				GC: "incorrect",
			},
			expectedError: compileopts.ErrGCInvalidOption,
		},
		{
			name: "GCOptionNone",
			opts: compileopts.Options{
				GC: "none",
			},
		},
		{
			name: "GCOptionLeaking",
			opts: compileopts.Options{
				GC: "leaking",
			},
		},
		{
			name: "GCOptionExtalloc",
			opts: compileopts.Options{
				GC: "extalloc",
			},
		},
		{
			name: "GCOptionConservative",
			opts: compileopts.Options{
				GC: "conservative",
			},
		},
		{
			name: "InvalidSchedulerOption",
			opts: compileopts.Options{
				Scheduler: "incorrect",
			},
			expectedError: compileopts.ErrSchedulerInvalidOption,
		},
		{
			name: "SchedulerOptionTasks",
			opts: compileopts.Options{
				Scheduler: "tasks",
			},
		},
		{
			name: "SchedulerOptionCoroutines",
			opts: compileopts.Options{
				Scheduler: "coroutines",
			},
		},
		{
			name: "InvalidPrintSizeOption",
			opts: compileopts.Options{
				PrintSizes: "invalid",
			},
			expectedError: compileopts.ErrPrintSizeInvalidOption,
		},
		{
			name: "PrintSizeOptionNone",
			opts: compileopts.Options{
				PrintSizes: "none",
			},
		},
		{
			name: "PrintSizeOptionShort",
			opts: compileopts.Options{
				PrintSizes: "short",
			},
		},
		{
			name: "PrintSizeOptionFull",
			opts: compileopts.Options{
				PrintSizes: "full",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Verify()
			if tc.expectedError != err {
				t.Errorf("expected %v, got %v", tc.expectedError, err)
			}
		})
	}
}
