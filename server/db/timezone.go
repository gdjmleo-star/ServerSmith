package db

import "time"

var TZ *time.Location

func init() {
	var err error
	TZ, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// fallback to UTC+8 fixed offset
		TZ = time.FixedZone("CST", 8*3600)
	}
}

// NowUTC returns current UTC time as RFC3339 string (for DB storage)
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// NowBeijing returns current Beijing time
func NowBeijing() time.Time {
	return time.Now().In(TZ)
}

// BeijingToUTC converts a Beijing-local time to UTC RFC3339 string
func BeijingToUTC(t time.Time) string {
	// Force interpret as Beijing if not already
	bj := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), TZ)
	return bj.UTC().Format(time.RFC3339)
}

// FormatUTCToBeijing formats a UTC time string to Beijing display string
func FormatUTCToBeijing(utcStr string) string {
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		return utcStr
	}
	return t.In(TZ).Format("2006-01-02 15:04:05")
}

// FormatUTCToBeijingShort formats UTC to Beijing date only
func FormatUTCToBeijingShort(utcStr string) string {
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		return utcStr
	}
	return t.In(TZ).Format("2006-01-02")
}
