package model

import (
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed sql/startup_cleanup_soft_deleted.sql
var startupCleanupSoftDeletedSQL string

// CleanupSoftDeletedRows 执行启动清理 SQL，移除历史软删除残留记录。
func CleanupSoftDeletedRows(db *gorm.DB) error {
	statements := splitSQLStatements(startupCleanupSoftDeletedSQL)
	if len(statements) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func splitSQLStatements(script string) []string {
	if strings.TrimSpace(script) == "" {
		return nil
	}

	lines := strings.Split(script, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		filtered = append(filtered, line)
	}

	joined := strings.Join(filtered, "\n")
	rawStatements := strings.Split(joined, ";")
	statements := make([]string, 0, len(rawStatements))
	for _, item := range rawStatements {
		stmt := strings.TrimSpace(item)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}
