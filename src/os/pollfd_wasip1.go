//go:build wasip1

package os

import "internal/poll"

// pollFD on wasip1 is the *poll.FD that backs net.FileListener /
// net.FileConn handoffs. The alias makes file.pfd directly typed as
// *poll.FD so PollFD reads/writes need no type conversion. The Exist
// method on *poll.FD (defined in internal/poll) absorbs the nil-check
// that file.close needs on the shared code path.
type pollFD = *poll.FD
