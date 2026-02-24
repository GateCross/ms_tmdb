package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cronMatcher struct {
	minutes [60]bool
	hours   [24]bool
	days    [31]bool
	months  [12]bool
	weeks   [7]bool
	domAny  bool
	dowAny  bool
}

func parseCronMatcher(expr string) (*cronMatcher, error) {
	raw := strings.TrimSpace(expr)
	if raw == "" {
		return nil, fmt.Errorf("cron 表达式不能为空")
	}

	parts := strings.Fields(raw)
	if len(parts) != 5 {
		return nil, fmt.Errorf("cron 表达式格式错误，应为 5 段：分 时 日 月 周")
	}

	minutes, _, err := parseCronField(parts[0], 0, 59, false)
	if err != nil {
		return nil, fmt.Errorf("分字段错误: %w", err)
	}
	hours, _, err := parseCronField(parts[1], 0, 23, false)
	if err != nil {
		return nil, fmt.Errorf("时字段错误: %w", err)
	}
	days, domAny, err := parseCronField(parts[2], 1, 31, false)
	if err != nil {
		return nil, fmt.Errorf("日字段错误: %w", err)
	}
	months, _, err := parseCronField(parts[3], 1, 12, false)
	if err != nil {
		return nil, fmt.Errorf("月字段错误: %w", err)
	}
	weeks, dowAny, err := parseCronField(parts[4], 0, 7, true)
	if err != nil {
		return nil, fmt.Errorf("周字段错误: %w", err)
	}

	matcher := &cronMatcher{
		domAny: domAny,
		dowAny: dowAny,
	}
	copy(matcher.minutes[:], minutes)
	copy(matcher.hours[:], hours)
	copy(matcher.days[:], days)
	copy(matcher.months[:], months)
	copy(matcher.weeks[:], weeks[:7])
	return matcher, nil
}

func (m *cronMatcher) Match(t time.Time) bool {
	month := int(t.Month())
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	week := int(t.Weekday())

	if !m.months[month-1] || !m.hours[hour] || !m.minutes[minute] {
		return false
	}

	domMatch := m.days[day-1]
	dowMatch := m.weeks[week]

	if m.domAny && m.dowAny {
		return true
	}
	if m.domAny {
		return dowMatch
	}
	if m.dowAny {
		return domMatch
	}
	return domMatch || dowMatch
}

func parseCronField(field string, min int, max int, isWeek bool) ([]bool, bool, error) {
	normalized := strings.TrimSpace(field)
	if normalized == "" {
		return nil, false, fmt.Errorf("字段为空")
	}

	size := max - min + 1
	if isWeek {
		size = 8
	}
	values := make([]bool, size)
	any := normalized == "*"

	tokens := strings.Split(normalized, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			return nil, false, fmt.Errorf("存在空片段")
		}
		if err := applyCronToken(values, token, min, max, isWeek); err != nil {
			return nil, false, err
		}
	}

	hasValue := false
	for i := range values {
		if values[i] {
			hasValue = true
			break
		}
	}
	if !hasValue {
		return nil, false, fmt.Errorf("字段未解析出有效值")
	}

	if isWeek && len(values) >= 8 && values[7] {
		values[0] = true
		values[7] = false
	}

	return values, any, nil
}

func applyCronToken(values []bool, token string, min int, max int, isWeek bool) error {
	base := token
	step := 1

	if strings.Contains(token, "/") {
		parts := strings.Split(token, "/")
		if len(parts) != 2 {
			return fmt.Errorf("步进语法错误: %s", token)
		}
		base = strings.TrimSpace(parts[0])
		parsedStep, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || parsedStep <= 0 {
			return fmt.Errorf("步进值非法: %s", token)
		}
		step = parsedStep
	}

	start, end, err := parseCronRange(base, min, max, isWeek)
	if err != nil {
		return err
	}
	if start > end {
		return fmt.Errorf("范围错误: %s", token)
	}

	for v := start; v <= end; v += step {
		index := v - min
		if isWeek && v == 7 {
			index = 7
		}
		if index < 0 || index >= len(values) {
			return fmt.Errorf("字段值越界: %d", v)
		}
		values[index] = true
	}

	return nil
}

func parseCronRange(base string, min int, max int, isWeek bool) (int, int, error) {
	if base == "*" || base == "" {
		return min, max, nil
	}

	if strings.Contains(base, "-") {
		parts := strings.Split(base, "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("范围语法错误: %s", base)
		}
		start, err := parseCronNumber(parts[0], min, max, isWeek)
		if err != nil {
			return 0, 0, err
		}
		end, err := parseCronNumber(parts[1], min, max, isWeek)
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	value, err := parseCronNumber(base, min, max, isWeek)
	if err != nil {
		return 0, 0, err
	}
	return value, value, nil
}

func parseCronNumber(raw string, min int, max int, isWeek bool) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("数值格式错误: %s", raw)
	}
	if isWeek && value == 7 {
		return 7, nil
	}
	if value < min || value > max {
		return 0, fmt.Errorf("值 %d 超出范围 [%d,%d]", value, min, max)
	}
	return value, nil
}
