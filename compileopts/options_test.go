package compileopts_test

import (
	"errors"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestVerifyOptions(t *testing.T) {

	expectedGCError := errors.New(`invalid gc option 'incorrect': valid values are none, leaking, conservative, custom`)
	expectedSchedulerError := errors.New(`invalid scheduler option 'incorrect': valid values are none, tasks, asyncify`)
	expectedPrintSizeError := errors.New(`invalid size option 'incorrect': valid values are none, short, full`)
	expectedPanicStrategyError := errors.New(`invalid panic option 'incorrect': valid values are print, trap`)

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
			expectedError: expectedGCError,
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
			name: "GCOptionConservative",
			opts: compileopts.Options{
				GC: "conservative",
			},
		},
		{
			name: "GCOptionCustom",
			opts: compileopts.Options{
				GC: "custom",
			},
		},
		{
			name: "InvalidSchedulerOption",
			opts: compileopts.Options{
				Scheduler: "incorrect",
			},
			expectedError: expectedSchedulerError,
		},
		{
			name: "SchedulerOptionNone",
			opts: compileopts.Options{
				Scheduler: "none",
			},
		},
		{
			name: "SchedulerOptionTasks",
			opts: compileopts.Options{
				Scheduler: "tasks",
			},
		},
		{
			name: "InvalidPrintSizeOption",
			opts: compileopts.Options{
				PrintSizes: "incorrect",
			},
			expectedError: expectedPrintSizeError,
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
		{
			name: "InvalidPanicOption",
			opts: compileopts.Options{
				PanicStrategy: "incorrect",
			},
			expectedError: expectedPanicStrategyError,
		},
		{
			name: "PanicOptionPrint",
			opts: compileopts.Options{
				PanicStrategy: "print",
			},
		},
		{
			name: "PanicOptionTrap",
			opts: compileopts.Options{
				PanicStrategy: "trap",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Verify()
			if tc.expectedError != err {
				if tc.expectedError.Error() != err.Error() {
					t.Errorf("expected %v, got %v", tc.expectedError, err)
				}
			}
		})
	}
}
