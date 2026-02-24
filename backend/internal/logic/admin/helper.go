package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/types"

	"gorm.io/gorm"
)

const (
	syncModeOverwriteAll    = "overwrite_all"
	syncModeUpdateUnchanged = "update_unmodified"
	syncModeSelective       = "selective"
	syncModePreview         = "preview"
)

// normalizePage 规范化分页参数
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

// normalizeSearchMode 规范化搜索模式，默认 contains
func normalizeSearchMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "prefix":
		return "prefix"
	default:
		return "contains"
	}
}

// applyKeywordFilter 在给定字段上应用模糊匹配
func applyKeywordFilter(db *gorm.DB, keyword, mode string, columns ...string) *gorm.DB {
	kw := strings.TrimSpace(keyword)
	if kw == "" || len(columns) == 0 {
		return db
	}

	pattern := "%" + kw + "%"
	if normalizeSearchMode(mode) == "prefix" {
		pattern = kw + "%"
	}

	parts := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s ILIKE ?", col))
		args = append(args, pattern)
	}

	return db.Where(strings.Join(parts, " OR "), args...)
}

// mergeTMDBData 把 patch 合并到原始 TMDB 数据中
func mergeTMDBData(base model.RawJSON, patch map[string]interface{}) (model.RawJSON, error) {
	if len(patch) == 0 {
		return base, nil
	}

	payload := make(map[string]interface{})
	if len(base) > 0 {
		if err := json.Unmarshal(base, &payload); err != nil {
			return nil, err
		}
	}

	for key, value := range patch {
		payload[key] = value
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return model.RawJSON(raw), nil
}

// toRawJSON 把任意对象转为 RawJSON
func toRawJSON(v interface{}) (model.RawJSON, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return model.RawJSON(raw), nil
}

// trimPtrString 去除字符串指针值的前后空格
func trimPtrString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

// buildGenresFromNames 把类型名转换为详情页可渲染结构
func buildGenresFromNames(names []string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	id := 1
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, map[string]interface{}{
			"id":   id,
			"name": name,
		})
		id++
	}
	return result
}

// normalizeSyncMode 标准化同步模式，默认更新未本地修改字段
func normalizeSyncMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case syncModeOverwriteAll:
		return syncModeOverwriteAll
	case syncModeSelective:
		return syncModeSelective
	case syncModePreview:
		return syncModePreview
	default:
		return syncModeUpdateUnchanged
	}
}

func rawJSONToMap(raw model.RawJSON) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if len(raw) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func cloneMap(input map[string]interface{}) map[string]interface{} {
	if len(input) == 0 {
		return map[string]interface{}{}
	}
	data, err := json.Marshal(input)
	if err != nil {
		output := make(map[string]interface{}, len(input))
		for k, v := range input {
			output[k] = v
		}
		return output
	}
	output := make(map[string]interface{})
	if err := json.Unmarshal(data, &output); err != nil {
		output = make(map[string]interface{}, len(input))
		for k, v := range input {
			output[k] = v
		}
	}
	return output
}

func mergeMap(base map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	result := cloneMap(base)
	for key, value := range patch {
		result[key] = value
	}
	return result
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sanitizeLocalPatch(localPatch map[string]interface{}, remote map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(localPatch))
	for key, value := range localPatch {
		remoteValue, exists := remote[key]
		if !exists || !reflect.DeepEqual(value, remoteValue) {
			result[key] = value
		}
	}
	return result
}

func removeFieldsFromPatch(localPatch map[string]interface{}, fields []string) map[string]interface{} {
	if len(localPatch) == 0 {
		return map[string]interface{}{}
	}

	overwriteSet := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		name := strings.TrimSpace(field)
		if name == "" {
			continue
		}
		overwriteSet[name] = struct{}{}
	}

	result := make(map[string]interface{}, len(localPatch))
	for key, value := range localPatch {
		if _, ok := overwriteSet[key]; ok {
			continue
		}
		result[key] = value
	}
	return result
}

// diffTopLevelFields 比较本地与远程顶层字段差异
func diffTopLevelFields(local map[string]interface{}, remote map[string]interface{}) []string {
	keySet := make(map[string]struct{}, len(local)+len(remote))
	for key := range local {
		keySet[key] = struct{}{}
	}
	for key := range remote {
		keySet[key] = struct{}{}
	}

	diff := make([]string, 0)
	for key := range keySet {
		localValue, localOK := local[key]
		remoteValue, remoteOK := remote[key]
		if !localOK || !remoteOK || !reflect.DeepEqual(localValue, remoteValue) {
			diff = append(diff, key)
		}
	}
	sort.Strings(diff)
	return diff
}

func filterDiffFieldsByLocalPatch(diffFields []string, localPatch map[string]interface{}) []string {
	if len(diffFields) == 0 || len(localPatch) == 0 {
		return diffFields
	}

	filtered := make([]string, 0, len(diffFields))
	for _, field := range diffFields {
		if _, ok := localPatch[field]; ok {
			continue
		}
		filtered = append(filtered, field)
	}
	return filtered
}

func splitDiffFieldsByLocalPatch(diffFields []string, localPatch map[string]interface{}, remote map[string]interface{}) ([]string, []string) {
	if len(diffFields) == 0 {
		return []string{}, []string{}
	}

	remoteDiff := make([]string, 0, len(diffFields))
	localOverrideDiff := make([]string, 0)
	for _, field := range diffFields {
		patchValue, ok := localPatch[field]
		if !ok {
			remoteDiff = append(remoteDiff, field)
			continue
		}

		if shouldIgnoreLocalOverrideDiff(field, patchValue, remote[field]) {
			continue
		}
		localOverrideDiff = append(localOverrideDiff, field)
	}
	return remoteDiff, localOverrideDiff
}

func shouldIgnoreLocalOverrideDiff(field string, patchValue interface{}, remoteValue interface{}) bool {
	return valuesEquivalent(field, patchValue, remoteValue)
}

func valuesEquivalent(field string, left interface{}, right interface{}) bool {
	switch field {
	case "genres":
		return equalGenresByName(left, right)
	case "credits", "combined_credits":
		return equalCreditsBySignature(left, right)
	case "production_companies":
		return equalObjectSliceByKeys(left, right, "id", "name")
	case "production_countries":
		return equalObjectSliceByKeys(left, right, "iso_3166_1", "name")
	case "spoken_languages":
		return equalObjectSliceByKeys(left, right, "iso_639_1", "english_name", "name")
	case "created_by":
		return equalObjectSliceByKeys(left, right, "id", "name")
	case "networks":
		return equalObjectSliceByKeys(left, right, "id", "name")
	case "seasons":
		return equalObjectSliceByKeys(left, right, "id", "season_number", "episode_count", "name")
	case "origin_country", "languages", "episode_run_time":
		return equalPrimitiveSlice(left, right)
	default:
		return false
	}
}

func equalGenresByName(patchValue interface{}, remoteValue interface{}) bool {
	return reflect.DeepEqual(normalizeGenreNames(patchValue), normalizeGenreNames(remoteValue))
}

func equalCreditsBySignature(left interface{}, right interface{}) bool {
	leftMap, leftOK := left.(map[string]interface{})
	rightMap, rightOK := right.(map[string]interface{})
	if !leftOK || !rightOK {
		return reflect.DeepEqual(left, right)
	}

	return reflect.DeepEqual(creditEntrySignatures(leftMap, "cast"), creditEntrySignatures(rightMap, "cast")) &&
		reflect.DeepEqual(creditEntrySignatures(leftMap, "crew"), creditEntrySignatures(rightMap, "crew"))
}

func creditEntrySignatures(payload map[string]interface{}, key string) []string {
	items, ok := payload[key].([]interface{})
	if !ok || len(items) == 0 {
		return []string{}
	}

	signatures := make([]string, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		signatures = append(signatures, buildCreditEntrySignature(entry))
	}
	sort.Strings(signatures)
	return signatures
}

func buildCreditEntrySignature(entry map[string]interface{}) string {
	keys := []string{
		"id",
		"credit_id",
		"name",
		"title",
		"original_name",
		"original_title",
		"character",
		"job",
		"department",
		"media_type",
		"episode_count",
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value, ok := entry[key]
		if !ok || value == nil {
			continue
		}

		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" {
			continue
		}
		parts = append(parts, key+"="+text)
	}

	if len(parts) == 0 {
		raw, err := json.Marshal(entry)
		if err != nil {
			return ""
		}
		return string(raw)
	}
	return strings.Join(parts, "|")
}

func equalObjectSliceByKeys(left interface{}, right interface{}, keys ...string) bool {
	leftItems, leftOK := left.([]interface{})
	rightItems, rightOK := right.([]interface{})
	if !leftOK || !rightOK {
		return reflect.DeepEqual(left, right)
	}
	return reflect.DeepEqual(objectSliceSignatures(leftItems, keys...), objectSliceSignatures(rightItems, keys...))
}

func objectSliceSignatures(items []interface{}, keys ...string) []string {
	if len(items) == 0 {
		return []string{}
	}

	signatures := make([]string, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			signatures = append(signatures, strings.TrimSpace(fmt.Sprintf("%v", item)))
			continue
		}

		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			value, ok := entry[key]
			if !ok || value == nil {
				continue
			}
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text == "" {
				continue
			}
			parts = append(parts, key+"="+text)
		}

		if len(parts) == 0 {
			raw, err := json.Marshal(entry)
			if err != nil {
				signatures = append(signatures, "")
			} else {
				signatures = append(signatures, string(raw))
			}
			continue
		}
		signatures = append(signatures, strings.Join(parts, "|"))
	}

	sort.Strings(signatures)
	return signatures
}

func equalPrimitiveSlice(left interface{}, right interface{}) bool {
	leftItems, leftOK := left.([]interface{})
	rightItems, rightOK := right.([]interface{})
	if !leftOK || !rightOK {
		return reflect.DeepEqual(left, right)
	}
	leftSignatures := primitiveSliceSignatures(leftItems)
	rightSignatures := primitiveSliceSignatures(rightItems)
	return reflect.DeepEqual(leftSignatures, rightSignatures)
}

func primitiveSliceSignatures(items []interface{}) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, strings.TrimSpace(fmt.Sprintf("%v", item)))
	}
	sort.Strings(result)
	return result
}

func filterEquivalentDiffFields(diffFields []string, local map[string]interface{}, remote map[string]interface{}) []string {
	if len(diffFields) == 0 {
		return diffFields
	}

	filtered := make([]string, 0, len(diffFields))
	for _, field := range diffFields {
		localValue, localOK := local[field]
		remoteValue, remoteOK := remote[field]
		if !localOK || !remoteOK {
			filtered = append(filtered, field)
			continue
		}
		if valuesEquivalent(field, localValue, remoteValue) {
			continue
		}
		filtered = append(filtered, field)
	}
	return filtered
}

func normalizeGenreNames(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}

	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(mapString(entry, "name")))
		if name == "" {
			continue
		}
		set[name] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for name := range set {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func filterIgnoredRemoteDiffFields(diffFields []string) []string {
	if len(diffFields) == 0 {
		return diffFields
	}

	ignored := map[string]struct{}{
		"images":              {},
		"videos":              {},
		"recommendations":     {},
		"similar":             {},
		"_ms_tv_season_local": {},
	}

	filtered := make([]string, 0, len(diffFields))
	for _, field := range diffFields {
		if _, ok := ignored[field]; ok {
			continue
		}
		filtered = append(filtered, field)
	}
	return filtered
}

func buildCompareDiffDetails(remoteDiffFields []string, localOverrideDiffFields []string, localData map[string]interface{}, localOverrideData map[string]interface{}, remoteData map[string]interface{}) []types.AdminCompareFieldDetail {
	details := make([]types.AdminCompareFieldDetail, 0, len(remoteDiffFields)+len(localOverrideDiffFields))
	details = appendCompareDiffDetails(details, remoteDiffFields, "remote", localData, remoteData)
	details = appendCompareDiffDetails(details, localOverrideDiffFields, "local_override", localOverrideData, remoteData)
	return details
}

func appendCompareDiffDetails(details []types.AdminCompareFieldDetail, fields []string, diffType string, localData map[string]interface{}, remoteData map[string]interface{}) []types.AdminCompareFieldDetail {
	for _, field := range fields {
		localValue, localOK := localData[field]
		remoteValue, remoteOK := remoteData[field]
		localText := "-"
		remoteText := "-"
		if localOK {
			localText = summarizeDiffValue(field, localValue)
		}
		if remoteOK {
			remoteText = summarizeDiffValue(field, remoteValue)
		}

		details = append(details, types.AdminCompareFieldDetail{
			Field:    field,
			DiffType: diffType,
			Local:    localText,
			Remote:   remoteText,
		})
	}
	return details
}

func summarizeDiffValue(field string, value interface{}) string {
	switch field {
	case "credits", "combined_credits":
		return summarizeCreditsValue(value)
	}

	switch v := value.(type) {
	case nil:
		return "-"
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return "-"
		}
		return truncateText(text, 80)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case []interface{}:
		return summarizeSliceValue(v)
	case map[string]interface{}:
		return summarizeMapValue(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return truncateText(fmt.Sprintf("%v", v), 80)
		}
		return truncateText(string(raw), 80)
	}
}

func summarizeCreditsValue(value interface{}) string {
	credits, ok := value.(map[string]interface{})
	if !ok {
		return summarizeDiffValue("", value)
	}

	castCount := mapArrayLen(credits, "cast")
	crewCount := mapArrayLen(credits, "crew")
	return fmt.Sprintf("cast:%d, crew:%d", castCount, crewCount)
}

func mapArrayLen(data map[string]interface{}, key string) int {
	items, ok := data[key].([]interface{})
	if !ok {
		return 0
	}
	return len(items)
}

func summarizeSliceValue(items []interface{}) string {
	if len(items) == 0 {
		return "0 项"
	}

	names := make([]string, 0, 3)
	for _, item := range items {
		name := pickDisplayName(item)
		if name == "" {
			continue
		}
		names = append(names, name)
		if len(names) >= 3 {
			break
		}
	}

	if len(names) == 0 {
		return fmt.Sprintf("%d 项", len(items))
	}
	return fmt.Sprintf("%s（共 %d 项）", strings.Join(names, "、"), len(items))
}

func summarizeMapValue(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	keys := sortedKeys(data)
	if len(keys) > 4 {
		keys = keys[:4]
	}

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := data[key]
		parts = append(parts, fmt.Sprintf("%s:%s", key, summarizeMapItemValue(value)))
	}
	return truncateText(strings.Join(parts, ", "), 80)
}

func summarizeMapItemValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "-"
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return "-"
		}
		return truncateText(text, 20)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case []interface{}:
		return fmt.Sprintf("%d项", len(v))
	case map[string]interface{}:
		return fmt.Sprintf("%d键", len(v))
	default:
		return truncateText(fmt.Sprintf("%v", v), 20)
	}
}

func pickDisplayName(value interface{}) string {
	entry, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	candidates := []string{"name", "title", "original_name", "original_title", "character", "job"}
	for _, key := range candidates {
		text := strings.TrimSpace(mapString(entry, key))
		if text != "" {
			return truncateText(text, 24)
		}
	}
	return ""
}

func truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func mapString(data map[string]interface{}, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func mapFloat64(data map[string]interface{}, key string) float64 {
	value, ok := data[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		return 0
	}
}

func mapInt(data map[string]interface{}, key string) int {
	value, ok := data[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(math.Trunc(v))
	case float32:
		return int(math.Trunc(float64(v)))
	case uint:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	default:
		return 0
	}
}

func mapInt64(data map[string]interface{}, key string) int64 {
	value, ok := data[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(math.Trunc(v))
	case float32:
		return int64(math.Trunc(float64(v)))
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	default:
		return 0
	}
}

func mapBool(data map[string]interface{}, key string) bool {
	value, ok := data[key]
	if !ok || value == nil {
		return false
	}
	v, ok := value.(bool)
	if !ok {
		return false
	}
	return v
}

func genreNamesFromRaw(raw model.RawJSON) []string {
	if len(raw) == 0 {
		return []string{}
	}

	payload := make(map[string]interface{})
	if err := json.Unmarshal(raw, &payload); err != nil {
		return []string{}
	}

	return extractGenreNamesFromMap(payload)
}

func extractGenreNamesFromMap(payload map[string]interface{}) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})

	appendName := func(raw string) {
		name := strings.TrimSpace(raw)
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		result = append(result, name)
	}

	if genres, ok := payload["genres"].([]interface{}); ok {
		for _, item := range genres {
			switch v := item.(type) {
			case map[string]interface{}:
				appendName(mapString(v, "name"))
			case string:
				appendName(v)
			}
		}
	}

	if names, ok := payload["genre_names"].([]interface{}); ok {
		for _, item := range names {
			if v, ok := item.(string); ok {
				appendName(v)
			}
		}
	}

	return result
}

func mergeGenreNames(primary []string, fallback []string) []string {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}

	result := make([]string, 0, len(primary)+len(fallback))
	seen := make(map[string]struct{}, len(primary)+len(fallback))
	appendName := func(raw string) {
		name := strings.TrimSpace(raw)
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, name)
	}

	for _, item := range primary {
		appendName(item)
	}
	for _, item := range fallback {
		appendName(item)
	}
	return result
}

func nextLocalTmdbID(db *gorm.DB, modelValue interface{}) (int, error) {
	var minID int
	if err := db.Model(modelValue).Select("COALESCE(MIN(tmdb_id), 0)").Where("tmdb_id < 0").Scan(&minID).Error; err != nil {
		return 0, err
	}
	if minID >= 0 {
		return -1, nil
	}
	return minID - 1, nil
}
