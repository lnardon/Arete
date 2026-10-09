// Package recurrence expands RFC 5545 recurrence rules the way Google
// Calendar does: a series row holds the rule lines, and its occurrences are
// computed for a requested range instead of being stored.
//
// Everything here is pure — no database, no HTTP — so the date math can be
// tested on its own.
package recurrence

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/teambition/rrule-go"

	"github.com/lnardon/arete/internal/models"
)

// MaxOccurrences caps how many occurrences of one series a single expansion
// returns, so a pathological rule can't blow up a response.
const MaxOccurrences = 1000

// ErrInvalid wraps every validation failure, so handlers can map it to a 400.
var ErrInvalid = errors.New("invalid recurrence")

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

var (
	utcUntilRe   = regexp.MustCompile(`^\d{8}T\d{6}Z$`)
	dateUntilRe  = regexp.MustCompile(`^\d{8}$`)
	instanceIDRe = regexp.MustCompile(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})_(\d{8}(?:T\d{6}Z)?)$`)
)

const (
	stampFormat     = "20060102T150405Z"
	dateStampFormat = "20060102"
)

var allowedFreqs = map[string]bool{"DAILY": true, "WEEKLY": true, "MONTHLY": true, "YEARLY": true}

// Validate checks recurrence lines sent by an Arete client. Lines pulled from
// Google skip it: they're Google's to define, and expansion copes with them.
func Validate(lines []string, allDay bool) error {
	if len(lines) == 0 {
		return nil
	}
	if len(lines) > 10 {
		return invalid("at most 10 recurrence lines are allowed")
	}

	rrules := 0
	for _, line := range lines {
		if len(line) > 500 {
			return invalid("recurrence lines must be 500 characters or fewer")
		}
		name, value, ok := splitLine(line)
		if !ok {
			return invalid("malformed recurrence line %q", line)
		}
		switch name {
		case "RRULE":
			rrules++
			if err := validateRule(value, allDay); err != nil {
				return err
			}
		case "EXRULE", "RDATE", "EXDATE":
		default:
			// Google rejects DTSTART/DTEND here too: the series start comes
			// from startAt/endAt.
			return invalid("unsupported recurrence line %s", name)
		}
	}
	if rrules != 1 {
		return invalid("exactly one RRULE line is required")
	}

	probe := models.CalendarEvent{
		StartAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:      time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		AllDay:     allDay,
		Timezone:   "UTC",
		Recurrence: lines,
	}
	if _, err := parse(probe); err != nil {
		return invalid("%v", err)
	}
	return nil
}

func validateRule(value string, allDay bool) error {
	parts, err := ruleParts(value)
	if err != nil {
		return invalid("%v", err)
	}
	if !allowedFreqs[parts["FREQ"]] {
		return invalid("FREQ must be DAILY, WEEKLY, MONTHLY or YEARLY")
	}
	if v, ok := parts["INTERVAL"]; ok {
		if n, err := strconv.Atoi(v); err != nil || n < 1 || n > 99 {
			return invalid("INTERVAL must be between 1 and 99")
		}
	}
	count, hasCount := parts["COUNT"]
	until, hasUntil := parts["UNTIL"]
	if hasCount && hasUntil {
		return invalid("COUNT and UNTIL can't both be set")
	}
	if hasCount {
		if n, err := strconv.Atoi(count); err != nil || n < 1 || n > 1000 {
			return invalid("COUNT must be between 1 and 1000")
		}
	}
	if hasUntil {
		// RFC 5545: UNTIL takes the same value type as DTSTART, and is UTC
		// when DTSTART carries a time zone.
		if allDay && !dateUntilRe.MatchString(until) {
			return invalid("UNTIL must be a date (YYYYMMDD) on an all-day event")
		}
		if !allDay && !utcUntilRe.MatchString(until) {
			return invalid("UNTIL must be a UTC date-time (YYYYMMDDTHHMMSSZ)")
		}
	}
	return nil
}

// splitLine splits "RRULE:FREQ=DAILY" or "EXDATE;TZID=X:2026..." into its
// upper-cased name and everything after the first ';' or ':'.
func splitLine(line string) (name, value string, ok bool) {
	line = strings.TrimSpace(line)
	i := strings.IndexAny(line, ";:")
	if i <= 0 || i == len(line)-1 {
		return "", "", false
	}
	return strings.ToUpper(line[:i]), line[i+1:], true
}

func ruleParts(value string) (map[string]string, error) {
	parts := map[string]string{}
	for _, part := range strings.Split(value, ";") {
		key, v, ok := strings.Cut(part, "=")
		if !ok || key == "" || v == "" {
			return nil, fmt.Errorf("malformed rule part %q", part)
		}
		key = strings.ToUpper(key)
		if _, dup := parts[key]; dup {
			return nil, fmt.Errorf("rule part %s is repeated", key)
		}
		parts[key] = v
	}
	return parts, nil
}

// withRulePart sets key=value in an RRULE value, replacing any existing
// part, or removes the part when value is empty. Other parts keep their
// original order and spelling, so a rule from Google round-trips unchanged.
func withRulePart(rule, key, value string) string {
	var out []string
	replaced := false
	for _, part := range strings.Split(rule, ";") {
		k, _, _ := strings.Cut(part, "=")
		if strings.EqualFold(k, key) {
			if value != "" && !replaced {
				out = append(out, key+"="+value)
				replaced = true
			}
			continue
		}
		out = append(out, part)
	}
	if value != "" && !replaced {
		out = append(out, key+"="+value)
	}
	return strings.Join(out, ";")
}

// ruleSet is a series' recurrence parsed against its own start. rrule-go's
// Set holds a single RRULE and ignores EXRULE, so the union and exclusions
// are done here instead.
type ruleSet struct {
	loc     *time.Location
	start   time.Time
	rules   []*rrule.RRule
	exrules []*rrule.RRule
	rdates  []time.Time
	exdates map[int64]bool
	finite  bool // every RRULE ends (COUNT or UNTIL)
}

func parse(series models.CalendarEvent) (*ruleSet, error) {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", series.Timezone, err)
	}
	return parseIn(series, loc)
}

// parseIn expands in the series' own zone so wall-clock times survive DST:
// a daily 8 p.m. event stays at 8 p.m. local when the UTC offset changes.
func parseIn(series models.CalendarEvent, loc *time.Location) (*ruleSet, error) {
	rs := &ruleSet{
		loc:     loc,
		start:   series.StartAt.In(loc).Truncate(time.Second),
		exdates: map[int64]bool{},
		finite:  true,
	}

	for _, line := range series.Recurrence {
		name, value, ok := splitLine(line)
		if !ok {
			return nil, fmt.Errorf("malformed recurrence line %q", line)
		}
		switch name {
		case "RRULE", "EXRULE":
			if !series.AllDay {
				value = untilAsEndOfDay(value)
			}
			opt, err := rrule.StrToROptionInLocation(value, loc)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", name, err)
			}
			opt.Dtstart = rs.start
			r, err := rrule.NewRRule(*opt)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", name, err)
			}
			if name == "RRULE" {
				rs.rules = append(rs.rules, r)
				if opt.Count == 0 && opt.Until.IsZero() {
					rs.finite = false
				}
			} else {
				rs.exrules = append(rs.exrules, r)
			}
		case "RDATE", "EXDATE":
			dates, err := rrule.StrToDatesInLoc(value, loc)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", name, err)
			}
			for _, d := range dates {
				if name == "RDATE" {
					rs.rdates = append(rs.rdates, d)
				} else {
					rs.exdates[d.Unix()] = true
				}
			}
		}
	}
	if len(rs.rules) == 0 {
		return nil, errors.New("no RRULE line")
	}
	return rs, nil
}

// untilAsEndOfDay widens a date-only UNTIL on a timed series to the end of
// that day, so the last day's occurrence is kept.
func untilAsEndOfDay(rule string) string {
	parts, err := ruleParts(rule)
	if err != nil {
		return rule
	}
	if until, ok := parts["UNTIL"]; ok && dateUntilRe.MatchString(until) {
		return withRulePart(rule, "UNTIL", until+"T235959")
	}
	return rule
}

// between returns the set's occurrence starts in [after, before], sorted.
func (rs *ruleSet) between(after, before time.Time) []time.Time {
	excluded := map[int64]bool{}
	for _, r := range rs.exrules {
		for _, t := range r.Between(after, before, true) {
			excluded[t.Unix()] = true
		}
	}

	seen := map[int64]bool{}
	var out []time.Time
	add := func(t time.Time) {
		if t.Before(after) || t.After(before) {
			return
		}
		k := t.Unix()
		if seen[k] || rs.exdates[k] || excluded[k] {
			return
		}
		seen[k] = true
		out = append(out, t.In(rs.loc))
	}

	// RFC 5545: DTSTART is always the first instance, even when the rule
	// itself wouldn't produce it. Google follows this.
	add(rs.start)
	for _, r := range rs.rules {
		for _, t := range r.Between(after, before, true) {
			add(t)
		}
	}
	for _, t := range rs.rdates {
		add(t)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// DayCount is an all-day span's length in whole days. Lengths are counted
// in days, not hours, because a day across a DST change is 23 or 25 hours.
func DayCount(start, end time.Time) int {
	days := int(math.Round(end.Sub(start).Hours() / 24))
	if days < 1 {
		days = 1
	}
	return days
}

func occurrenceEnd(series models.CalendarEvent, loc *time.Location, start time.Time) time.Time {
	if series.AllDay {
		return start.In(loc).AddDate(0, 0, DayCount(series.StartAt, series.EndAt))
	}
	return start.Add(series.EndAt.Sub(series.StartAt))
}

// OccurrenceEnd is when the occurrence of series starting at start ends.
func OccurrenceEnd(series models.CalendarEvent, start time.Time) time.Time {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return occurrenceEnd(series, loc, start)
}

// EndFrom gives an occurrence moved to newStart the same length as the span
// [start, end): days for all-day events, a duration otherwise.
func EndFrom(newStart, start, end time.Time, allDay bool, loc *time.Location) time.Time {
	if allDay {
		return newStart.In(loc).AddDate(0, 0, DayCount(start, end))
	}
	return newStart.Add(end.Sub(start))
}

func expand(series models.CalendarEvent, loc *time.Location, from, to time.Time) ([]time.Time, error) {
	rs, err := parseIn(series, loc)
	if err != nil {
		return nil, err
	}

	// An occurrence overlapping [from, to) can start up to one occurrence
	// length before from; the extra hour covers all-day lengths across DST.
	lookback := series.EndAt.Sub(series.StartAt) + time.Hour
	var out []time.Time
	for _, start := range rs.between(from.Add(-lookback), to) {
		if len(out) == MaxOccurrences {
			break
		}
		if start.Before(to) && occurrenceEnd(series, loc, start).After(from) {
			out = append(out, start)
		}
	}
	return out, nil
}

// Expand returns the original start times of the series' occurrences that
// overlap [from, to), before any exceptions are applied.
func Expand(series models.CalendarEvent, from, to time.Time) ([]time.Time, error) {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", series.Timezone, err)
	}
	return expand(series, loc, from, to)
}

// End returns when the series' last occurrence ends, or nil when the series
// repeats forever. It's stored as recurrence_end_at for range queries.
func End(series models.CalendarEvent) (*time.Time, error) {
	rs, err := parse(series)
	if err != nil {
		return nil, err
	}
	if !rs.finite {
		return nil, nil
	}

	last := rs.start
	for _, r := range rs.rules {
		if all := r.All(); len(all) > 0 && all[len(all)-1].After(last) {
			last = all[len(all)-1]
		}
	}
	for _, d := range rs.rdates {
		if d.After(last) {
			last = d
		}
	}
	end := occurrenceEnd(series, rs.loc, last)
	return &end, nil
}

// IsOccurrence reports whether t is one of the series' occurrence starts.
func IsOccurrence(series models.CalendarEvent, t time.Time) bool {
	rs, err := parse(series)
	if err != nil {
		return false
	}
	for _, start := range rs.between(t, t) {
		if start.Equal(t) {
			return true
		}
	}
	return false
}

// TrimBefore ends the series just before splitAt, for "this and following"
// edits and deletes. It returns the rewritten lines and how many occurrences
// remain before splitAt. COUNT is replaced by UNTIL, since RFC 5545 forbids
// both on one rule.
func TrimBefore(series models.CalendarEvent, splitAt time.Time) (lines []string, kept int, err error) {
	rs, err := parse(series)
	if err != nil {
		return nil, 0, err
	}
	kept = len(rs.between(rs.start, splitAt.Add(-time.Second)))

	until := splitAt.Add(-time.Second).UTC().Format(stampFormat)
	if series.AllDay {
		until = splitAt.In(rs.loc).AddDate(0, 0, -1).Format(dateStampFormat)
	}
	for _, line := range series.Recurrence {
		if name, value, ok := splitLine(line); ok && name == "RRULE" {
			value = withRulePart(withRulePart(value, "COUNT", ""), "UNTIL", until)
			line = "RRULE:" + value
		}
		lines = append(lines, line)
	}
	return lines, kept, nil
}

// ReduceCount lowers COUNT on every RRULE by n (never below 1), for the new
// series of a split that keeps the original rule: the occurrences already
// spent on the old series no longer count toward it.
func ReduceCount(lines []string, n int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if name, value, ok := splitLine(line); ok && name == "RRULE" {
			if parts, err := ruleParts(value); err == nil {
				if count, err := strconv.Atoi(parts["COUNT"]); err == nil {
					line = "RRULE:" + withRulePart(value, "COUNT", strconv.Itoa(max(1, count-n)))
				}
			}
		}
		out = append(out, line)
	}
	return out
}

// SameLines reports whether two recurrences are identical, line for line.
func SameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func instanceID(seriesID string, start time.Time, allDay bool, loc *time.Location) string {
	if allDay {
		return seriesID + "_" + start.In(loc).Format(dateStampFormat)
	}
	return seriesID + "_" + start.UTC().Format(stampFormat)
}

// InstanceID names one occurrence the way Google does:
// <seriesId>_20261009T210000Z for timed series (UTC), <seriesId>_20261009 for
// all-day ones (in the series' zone). It identifies the original slot, so it
// stays the same when the occurrence is edited or moved.
func InstanceID(series models.CalendarEvent, originalStart time.Time) string {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return instanceID(series.ID, originalStart, series.AllDay, loc)
}

// ParseInstanceID splits an instance ID into its series ID and stamp. ok is
// false for anything else, including a plain event UUID.
func ParseInstanceID(id string) (seriesID, stamp string, ok bool) {
	m := instanceIDRe.FindStringSubmatch(id)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

// ParseStamp turns an instance ID's stamp back into the original start.
func ParseStamp(series models.CalendarEvent, stamp string) (time.Time, error) {
	if len(stamp) == len(dateStampFormat) {
		loc, err := time.LoadLocation(series.Timezone)
		if err != nil {
			return time.Time{}, err
		}
		return time.ParseInLocation(dateStampFormat, stamp, loc)
	}
	return time.Parse(stampFormat, stamp)
}

// Shift carries an edit of one occurrence over to the rest of a series. The
// edited occurrence moved from `from` to `to`; every other slot moves by the
// same number of calendar days and the same wall-clock offset, so a
// 6 p.m. → 7 p.m. edit keeps every occurrence on its own date, at 7 p.m.
type Shift struct {
	fromLoc, toLoc *time.Location
	days           int
	clock          time.Duration
}

func NewShift(from time.Time, fromLoc *time.Location, to time.Time, toLoc *time.Location) Shift {
	f, t := from.In(fromLoc), to.In(toLoc)
	fy, fm, fd := f.Date()
	ty, tm, td := t.Date()
	days := int(math.Round(time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC).Sub(time.Date(fy, fm, fd, 0, 0, 0, 0, time.UTC)).Hours() / 24))
	return Shift{fromLoc: fromLoc, toLoc: toLoc, days: days, clock: clockOf(t) - clockOf(f)}
}

func clockOf(t time.Time) time.Duration {
	h, m, s := t.Clock()
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
}

// Apply moves t the way the edited occurrence moved. time.Date normalizes a
// clock that runs past midnight and resolves the wall time in the target
// zone, so DST is respected.
func (s Shift) Apply(t time.Time) time.Time {
	l := t.In(s.fromLoc)
	y, m, d := l.Date()
	return time.Date(y, m, d+s.days, 0, 0, int((clockOf(l)+s.clock)/time.Second), 0, s.toLoc)
}
