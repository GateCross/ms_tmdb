-- 清理历史软删除残留数据（启动时自动执行）
DELETE FROM movie_lang_snapshots WHERE deleted_at IS NOT NULL;
DELETE FROM tv_lang_snapshots WHERE deleted_at IS NOT NULL;
DELETE FROM person_lang_snapshots WHERE deleted_at IS NOT NULL;
DELETE FROM movies WHERE deleted_at IS NOT NULL;
DELETE FROM tv_series WHERE deleted_at IS NOT NULL;
DELETE FROM people WHERE deleted_at IS NOT NULL;
