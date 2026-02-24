package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAutoSyncSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAutoSyncSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAutoSyncSettingsLogic {
	return &UpdateAutoSyncSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAutoSyncSettingsLogic) UpdateAutoSyncSettings(req *types.AdminAutoSyncReq) (*types.AdminAutoSyncResp, error) {
	current := normalizeAutoSyncSettings(AutoSyncSettings{
		Enabled:          l.svcCtx.Config.Tmdb.AutoSync.Enabled,
		CronExpr:         l.svcCtx.Config.Tmdb.AutoSync.CronExpr,
		Mode:             l.svcCtx.Config.Tmdb.AutoSync.Mode,
		BatchSize:        l.svcCtx.Config.Tmdb.AutoSync.BatchSize,
		StartDelaySecond: l.svcCtx.Config.Tmdb.AutoSync.StartDelaySecond,
	})

	scheduler := GetLibraryAutoSyncScheduler()
	if scheduler != nil {
		current = scheduler.GetSettings()
	}

	next := current
	if req.Enabled != nil {
		next.Enabled = *req.Enabled
	}
	if strings.TrimSpace(req.CronExpr) != "" {
		next.CronExpr = req.CronExpr
	}
	if strings.TrimSpace(req.Mode) != "" {
		next.Mode = req.Mode
	}
	if req.BatchSize != nil {
		next.BatchSize = *req.BatchSize
	}
	if req.StartDelaySecond != nil {
		next.StartDelaySecond = *req.StartDelaySecond
	}
	next = normalizeAutoSyncSettings(next)

	if scheduler != nil {
		var updateErr error
		next, updateErr = scheduler.UpdateSettings(next)
		if updateErr != nil {
			return nil, updateErr
		}
	}

	oldConfig := l.svcCtx.Config.Tmdb.AutoSync
	applyAutoSyncConfigToServiceContext(l.svcCtx, next)

	configFile := strings.TrimSpace(l.svcCtx.Config.ConfigFile)
	if configFile == "" {
		configFile = "etc/tmdb.yaml"
	}
	if err := writeAutoSyncToConfigFile(configFile, next); err != nil {
		if scheduler != nil {
			_, _ = scheduler.UpdateSettings(current)
		}
		l.svcCtx.Config.Tmdb.AutoSync = oldConfig
		return nil, err
	}

	return &types.AdminAutoSyncResp{
		Enabled:          next.Enabled,
		CronExpr:         next.CronExpr,
		Mode:             next.Mode,
		BatchSize:        next.BatchSize,
		StartDelaySecond: next.StartDelaySecond,
		Running:          next.Running,
	}, nil
}

func applyAutoSyncConfigToServiceContext(svcCtx *svc.ServiceContext, settings AutoSyncSettings) {
	svcCtx.Config.Tmdb.AutoSync.Enabled = settings.Enabled
	svcCtx.Config.Tmdb.AutoSync.CronExpr = settings.CronExpr
	svcCtx.Config.Tmdb.AutoSync.Mode = settings.Mode
	svcCtx.Config.Tmdb.AutoSync.BatchSize = settings.BatchSize
	svcCtx.Config.Tmdb.AutoSync.StartDelaySecond = settings.StartDelaySecond
}

func writeAutoSyncToConfigFile(configPath string, settings AutoSyncSettings) error {
	configFilePath := filepath.Clean(strings.TrimSpace(configPath))
	if configFilePath == "" {
		return errors.New("配置文件路径为空")
	}

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	raw := string(content)
	lineSep := "\n"
	if strings.Contains(raw, "\r\n") {
		lineSep = "\r\n"
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
	}

	lines := strings.Split(raw, "\n")
	tmdbFound := false
	tmdbStart := -1
	tmdbIndent := 0
	tmdbEnd := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := leadingIndentLen(line)

		if !tmdbFound {
			if strings.HasPrefix(trimmed, "Tmdb:") {
				tmdbFound = true
				tmdbStart = i
				tmdbIndent = indent
			}
			continue
		}

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && indent <= tmdbIndent {
			tmdbEnd = i
			break
		}
	}
	if !tmdbFound {
		return errors.New("配置文件缺少 Tmdb 段")
	}

	parentIndent := tmdbIndent + 2
	autoSyncStart := -1
	autoSyncEnd := tmdbEnd
	for i := tmdbStart + 1; i < tmdbEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		indent := leadingIndentLen(lines[i])

		if strings.HasPrefix(trimmed, "AutoSync:") && indent == parentIndent {
			autoSyncStart = i
			autoSyncEnd = tmdbEnd
			for j := i + 1; j < tmdbEnd; j++ {
				childTrimmed := strings.TrimSpace(lines[j])
				childIndent := leadingIndentLen(lines[j])
				if childTrimmed != "" && !strings.HasPrefix(childTrimmed, "#") && childIndent <= parentIndent {
					autoSyncEnd = j
					break
				}
			}
			break
		}
	}

	indent := strings.Repeat(" ", parentIndent)
	childIndent := strings.Repeat(" ", parentIndent+2)
	block := []string{
		indent + "AutoSync:",
		childIndent + "Enabled: " + strconv.FormatBool(settings.Enabled),
		childIndent + "CronExpr: " + yamlDoubleQuoted(settings.CronExpr),
		childIndent + "Mode: " + yamlDoubleQuoted(settings.Mode),
		childIndent + "BatchSize: " + strconv.Itoa(settings.BatchSize),
		childIndent + "StartDelaySecond: " + strconv.Itoa(settings.StartDelaySecond),
	}

	if autoSyncStart >= 0 {
		lines = append(lines[:autoSyncStart], append(block, lines[autoSyncEnd:]...)...)
	} else {
		lines = append(lines[:tmdbEnd], append(block, lines[tmdbEnd:]...)...)
	}

	output := strings.Join(lines, "\n")
	if lineSep == "\r\n" {
		output = strings.ReplaceAll(output, "\n", "\r\n")
	}

	if err := os.WriteFile(configFilePath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}
