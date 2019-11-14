package time

type Zone struct {
	Name   string // Abbreviated name ("CET", "CEST").
	Offset int    // Seconds east of UTC.
}

// DST describes daylight saving time zone. 25 least significant bits of Start
// and End contain seconds from begining of the year to the month-weekday-hour
// at which the DST starts/ends, assuming that the year is not a leap year and
// its first day is Monday. 7 most significant bits of Start and End contain
// margin, a number of days that weekdays can be shifted back to do not
// introduce new last weekday at end of month or to do not lose first weekday
// at beginning of month. Examples of not a leap years whose first day is
// Monday: 2018, 2029.
type DST struct {
	Zone  *Zone
	Start uint32
	End   uint32
}

// A Location maps time instants to the zone in use at that time.
// This is simplified implementation that does not support historical changes.
type Location struct {
	Name string
	Zone *Zone
	DST  *DST // Nil if DST not used in location.
}

func (l *Location) String() string {
	return l.Name
}

var utcLoc = Location{"UTC", &Zone{"UTC", 0}, nil}

// UTC represents Universal Coordinated Time (UTC).
var UTC *Location = &utcLoc

// Local represents the system's local time zone.
var Local = &utcLoc

// FixedZone returns a Location that always uses
// the given zone name and offset (seconds east of UTC).
func FixedZone(name string, offset int) *Location {
	return &Location{Name: name, Zone: &Zone{name, offset}}
}

// Lookup returns information about the time zone in use at an instant in time
// expressed as absolute time abs. The returned information gives the name of
// the zone (such as "CET"), the offset in seconds east of UTC, the start and
// end times bracketing abs when that zone is in effect. If start/end falls on
// the previous or next year, the approximate value of start/end is returned.
// For now only Date uses these values and works fine with such approximation.
func (l *Location) lookup(abs uint64) (name string, offset int, start, end uint64) {
	if l == nil {
		l = &utcLoc
	}
	if l.DST == nil {
		return l.Zone.Name, l.Zone.Offset, 0, 1<<64 - 1
	}

	// This code is similar to the code of absDate. See absDate for better
	// description of any step.

	// Avoid 64-bit calculations.

	// Second of 400-year cycle.
	s400 := abs % (daysPer400Years * secondsPerDay)

	// Day of 400-year cycle.
	d400 := int(s400 / secondsPerDay)

	// Second of day.
	s := int(s400 - uint64(d400)*secondsPerDay)

	// Day of 100-year cycle.
	n100 := d400 / daysPer100Years
	n100 -= n100 >> 2
	d := d400 - daysPer100Years*n100

	// Day of 4-year cycle.
	n4 := d / daysPer4Years
	d -= daysPer4Years * n4

	// Day of year (0 means first day).
	n := d / 365
	n -= n >> 2
	d -= 365 * n

	// Calculate second of year and determine does the year is a leap year.
	ys := d*secondsPerDay + s
	isLeap := (n == 4-1 && (n4 != 25-1 || n100 == 4-1))

	// Weekday of first year day.
	wday := (d400 - d) % 7 // Zero means Monday.

	// Adjust l.DST.Start and l.DST.End that they describe always the same time
	// on the same month and the same weakday.
	dstStart, margin := int(l.DST.Start&0x1FFFFFF), int(l.DST.Start>>25)
	adj := wday
	if isLeap && dstStart > (31+28+15)*secondsPerDay {
		// BUG: dstStart > (31+28+15)*secondsPerDay is simplified condition.
		// Correct condition should use direction bit of margin (not
		// implemented) to detect that margin describes first n-th weekday
		// (Saturday, Sunday) of March or last n-th weekday of March.
		margin--
	}
	if wday >= margin {
		adj -= 7
	}
	dstStart -= adj * secondsPerDay
	dstEnd, margin := int(l.DST.End&0x1FFFFFF), int(l.DST.End>>25)
	adj = wday
	if isLeap && dstEnd > (31+28+15)*secondsPerDay {
		// BUG: See above.
		margin--
	}
	if wday >= margin {
		adj -= 7
	}
	dstEnd -= adj * secondsPerDay

	// Date function wants to classify ambiguous time period as not DST
	if change := l.DST.Zone.Offset - l.Zone.Offset; change > 0 {
		dstEnd -= change
	} else {
		dstStart += change
	}

	abs -= uint64(ys)              // beginning of year
	start = abs + uint64(dstStart) // start of DST
	end = abs + uint64(dstEnd)     // end of DST

	if dstStart < dstEnd {
		if ys < dstStart {
			return l.Zone.Name, l.Zone.Offset, end - 365*secondsPerDay, start
		}
		if dstEnd <= ys {
			return l.Zone.Name, l.Zone.Offset, end, start + 365*secondsPerDay
		}
		return l.DST.Zone.Name, l.DST.Zone.Offset, start, end
	}
	if ys < dstEnd {
		return l.DST.Zone.Name, l.DST.Zone.Offset, start - 365*secondsPerDay, end
	}
	if dstStart <= ys {
		return l.DST.Zone.Name, l.DST.Zone.Offset, start, end + 365*secondsPerDay
	}

	return l.Zone.Name, l.Zone.Offset, end, start
}

// lookupName returns information about the time zone with
// the given name (such as "EST") at the given absolute time
// (what the given time of day would be in UTC).
func (l *Location) lookupName(name string, abs uint64) (offset int, ok bool) {
	zones := [2]*Zone{l.Zone, l.DST.Zone}
	for _, zone := range zones {
		if zone == nil {
			break
		}
		if zone.Name == name {
			nam, offset, _, _ := l.lookup(abs - uint64(zone.Offset))
			if nam == zone.Name {
				return offset, true
			}
		}
	}
	// fall back to an ordinary name match.
	for _, zone := range zones {
		if zone == nil {
			break
		}
		if zone.Name == name {
			return zone.Offset, true
		}
	}
	return
}
