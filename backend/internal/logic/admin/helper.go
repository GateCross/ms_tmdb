package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"ms_tmdb/internal/model"

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
