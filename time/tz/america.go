// Copyright 2019 Michal Derkacz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tz

import "github.com/embeddedgo/time"

var (
	EST = time.Zone{"EST", -5 * 3600}
	EDT = time.DST{
		// Second Sunday in March 2:00 to first Sunday in November 3:00.
		&time.Zone{"EDT", -4 * 3600},
		// Start: March 11 7:00 UTC, first Sunday this month is March 4.
		(69*24+7)*3600 | 4<<25,
		// End: November 4 7:00 UTC, first Sunday this month is November 4.
		(307*24+7)*3600 | 4<<25,
	}
	PST = time.Zone{"PST", -8 * 3600}
	PDT = time.DST{
		// Second Sunday in March 2:00 to first Sunday in November 3:00.
		&time.Zone{"PDT", -7 * 3600},
		// Start: March 11 10:00 UTC, first Sunday this month is March 4.
		(69*24+10)*3600 | 4<<25,
		// End: November 4 10:00 UTC, first Sunday this month is November 4.
		(307*24+10)*3600 | 4<<25,
	}
)

var (
	AmericaNewYork    = time.Location{"America/New_York", &EST, &EDT}
	AmericaLosAngeles = time.Location{"America/Los_Angeles", &PST, &PDT}
)
