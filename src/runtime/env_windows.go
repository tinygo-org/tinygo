//go:build windows

package runtime

// Set environment variable in Windows:
//
//	BOOL SetEnvironmentVariableA(
//	    [in]           LPCSTR lpName,
//	    [in, optional] LPCSTR lpValue
//	);
//
//export SetEnvironmentVariableA
func _SetEnvironmentVariableA(key, val *byte) bool

func setenv(key, val *byte) {
	// ignore any errors
	_SetEnvironmentVariableA(key, val)
}

func unsetenv(key *byte) {
	// ignore any errors
	_SetEnvironmentVariableA(key, nil)
}
