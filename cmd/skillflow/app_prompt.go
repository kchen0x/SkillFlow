package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	promptcatalogapp "github.com/shinerio/skillflow/core/promptcatalog/app"
	promptdomain "github.com/shinerio/skillflow/core/promptcatalog/domain"
	promptrepo "github.com/shinerio/skillflow/core/promptcatalog/infra/repository"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) promptService() (*promptcatalogapp.Service, string, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, "", err
	}
	root := filepath.Join(a.backupRootDir(cfg), "prompts")
	return promptcatalogapp.NewService(promptrepo.NewFilesystemStorage(root)), root, nil
}

func (a *App) ListPrompts() ([]*promptdomain.Prompt, error) {
	service, _, err := a.promptService()
	if err != nil {
		return nil, err
	}
	return service.ListPrompts()
}

func (a *App) ListPromptCategories() ([]string, error) {
	service, _, err := a.promptService()
	if err != nil {
		return nil, err
	}
	categories, err := service.ListPromptCategories()
	if err != nil {
		return nil, err
	}
	hasDefault := false
	for _, category := range categories {
		if category == promptdomain.DefaultCategoryName {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		categories = append([]string{promptdomain.DefaultCategoryName}, categories...)
	}
	return categories, nil
}

func (a *App) CreatePrompt(name, description, category, content string, imageURLs []string, webLinksMarkdown string) (*promptdomain.Prompt, error) {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt create failed: load storage failed: %v", err)
		return nil, err
	}
	webLinks, err := promptcatalogapp.ParseWebLinksMarkdown(webLinksMarkdown)
	if err != nil {
		a.logErrorf("prompt create failed: prompt=%s category=%s root=%s err=%v", name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt create started: prompt=%s category=%s images=%d links=%d root=%s", name, category, len(imageURLs), len(webLinks), root)
	item, err := service.CreatePrompt(name, description, category, content, imageURLs, webLinks)
	if err != nil {
		a.logErrorf("prompt create failed: prompt=%s category=%s root=%s err=%v", name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt create completed: prompt=%s category=%s file=%s", item.Name, item.Category, item.FilePath)
	a.scheduleAutoBackup()
	return item, nil
}

func (a *App) UpdatePrompt(originalName, name, description, category, content string, imageURLs []string, webLinksMarkdown string) (*promptdomain.Prompt, error) {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt update failed: prompt=%s load storage failed: %v", originalName, err)
		return nil, err
	}
	webLinks, err := promptcatalogapp.ParseWebLinksMarkdown(webLinksMarkdown)
	if err != nil {
		a.logErrorf("prompt update failed: prompt=%s next=%s category=%s root=%s err=%v", originalName, name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt update started: prompt=%s next=%s category=%s images=%d links=%d root=%s", originalName, name, category, len(imageURLs), len(webLinks), root)
	item, err := service.UpdatePrompt(originalName, name, description, category, content, imageURLs, webLinks)
	if err != nil {
		a.logErrorf("prompt update failed: prompt=%s next=%s category=%s root=%s err=%v", originalName, name, category, root, err)
		return nil, err
	}
	a.logInfof("prompt update completed: prompt=%s category=%s file=%s", item.Name, item.Category, item.FilePath)
	a.scheduleAutoBackup()
	return item, nil
}

func (a *App) MovePromptCategory(name string, category string) error {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt move category failed: prompt=%s category=%s load storage failed: %v", name, category, err)
		return err
	}
	a.logInfof("prompt move category started: prompt=%s category=%s root=%s", name, category, root)
	if err := service.MovePromptToCategory(name, category); err != nil {
		a.logErrorf("prompt move category failed: prompt=%s category=%s root=%s err=%v", name, category, root, err)
		return err
	}
	a.logInfof("prompt move category completed: prompt=%s category=%s", name, category)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) DeletePrompt(name string) error {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt delete failed: prompt=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt delete started: prompt=%s root=%s", name, root)
	if err := service.DeletePrompt(name); err != nil {
		a.logErrorf("prompt delete failed: prompt=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt delete completed: prompt=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) CreatePromptCategory(name string) error {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt category create failed: category=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt category create started: category=%s root=%s", name, root)
	if err := service.CreatePromptCategory(name); err != nil {
		a.logErrorf("prompt category create failed: category=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt category create completed: category=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) RenamePromptCategory(oldName, newName string) error {
	if oldName == promptdomain.DefaultCategoryName {
		return fmt.Errorf("默认分类不可重命名")
	}
	if newName == promptdomain.DefaultCategoryName {
		return fmt.Errorf("不能重命名为默认分类")
	}
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt category rename failed: category=%s next=%s load storage failed: %v", oldName, newName, err)
		return err
	}
	a.logInfof("prompt category rename started: category=%s next=%s root=%s", oldName, newName, root)
	if err := service.RenamePromptCategory(oldName, newName); err != nil {
		a.logErrorf("prompt category rename failed: category=%s next=%s root=%s err=%v", oldName, newName, root, err)
		return err
	}
	a.logInfof("prompt category rename completed: category=%s next=%s", oldName, newName)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) DeletePromptCategory(name string) error {
	if name == promptdomain.DefaultCategoryName {
		return fmt.Errorf("默认分类不可删除")
	}
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt category delete failed: category=%s load storage failed: %v", name, err)
		return err
	}
	a.logInfof("prompt category delete started: category=%s root=%s", name, root)
	if err := service.DeletePromptCategory(name); err != nil {
		a.logErrorf("prompt category delete failed: category=%s root=%s err=%v", name, root, err)
		return err
	}
	a.logInfof("prompt category delete completed: category=%s", name)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) ImportPrompts() (int, error) {
	result, err := a.PrepareImportPrompts()
	if err != nil {
		return 0, err
	}
	if result == nil || result.SessionID == "" {
		return 0, nil
	}
	return a.CompleteImportPrompts(result.SessionID, promptImportNames(result.Conflicts))
}

func (a *App) PrepareImportPrompts() (*PromptImportPrepareResult, error) {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt import prepare failed: load storage failed: %v", err)
		return nil, err
	}
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "导入提示词",
		DefaultDirectory: nearestExistingDirectory(root),
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON Files (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		a.logErrorf("prompt import prepare failed: open dialog err=%v", err)
		return nil, err
	}
	if filePath == "" {
		return &PromptImportPrepareResult{}, nil
	}
	a.logInfof("prompt import prepare started: file=%s", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		a.logErrorf("prompt import prepare failed: file=%s err=%v", filePath, err)
		return nil, err
	}
	preview, err := service.PreviewPromptImport(data)
	if err != nil {
		a.logErrorf("prompt import prepare failed: file=%s err=%v", filePath, err)
		return nil, err
	}
	sessionID := a.promptImports.Create(filePath, preview)
	a.logInfof("prompt import prepare completed: file=%s session=%s creates=%d conflicts=%d", filePath, sessionID, len(preview.Creates), len(preview.Conflicts))
	return &PromptImportPrepareResult{
		SessionID: sessionID,
		Creates:   preview.Creates,
		Conflicts: preview.Conflicts,
	}, nil
}

func (a *App) CompleteImportPrompts(sessionID string, overwriteNames []string) (int, error) {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt import complete failed: session=%s load storage failed: %v", sessionID, err)
		return 0, err
	}
	session, ok := a.promptImports.Take(sessionID)
	if !ok {
		err := fmt.Errorf("prompt import session not found")
		a.logErrorf("prompt import complete failed: session=%s root=%s err=%v", sessionID, root, err)
		return 0, err
	}
	a.logInfof("prompt import complete started: session=%s file=%s creates=%d conflicts=%d overwrites=%d root=%s", sessionID, session.FilePath, len(session.Preview.Creates), len(session.Preview.Conflicts), len(overwriteNames), root)
	count, err := service.ApplyPromptImport(session.Preview, overwriteNames)
	if err != nil {
		a.logErrorf("prompt import complete failed: session=%s file=%s root=%s err=%v", sessionID, session.FilePath, root, err)
		return 0, err
	}
	a.logInfof("prompt import complete completed: session=%s file=%s count=%d", sessionID, session.FilePath, count)
	a.scheduleAutoBackup()
	return count, nil
}

func (a *App) CancelImportPrompts(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	a.promptImports.Delete(sessionID)
	a.logInfof("prompt import cancel completed: session=%s", sessionID)
	return nil
}

func (a *App) ExportPrompts() (string, error) {
	return a.ExportPromptsByNames(nil)
}

func (a *App) ExportPromptsByNames(names []string) (string, error) {
	service, root, err := a.promptService()
	if err != nil {
		a.logErrorf("prompt export failed: load storage failed: %v", err)
		return "", err
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "导出提示词",
		DefaultDirectory:     nearestExistingDirectory(root),
		DefaultFilename:      fmt.Sprintf("skillflow-prompts-%s.json", time.Now().Format("20060102-150405")),
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON Files (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		a.logErrorf("prompt export failed: save dialog err=%v", err)
		return "", err
	}
	if filePath == "" {
		return "", nil
	}
	a.logInfof("prompt export started: file=%s promptCount=%d", filePath, len(names))
	data, err := service.ExportPromptBundle(names)
	if err != nil {
		a.logErrorf("prompt export failed: file=%s err=%v", filePath, err)
		return "", err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		a.logErrorf("prompt export failed: file=%s err=%v", filePath, err)
		return "", err
	}
	a.logInfof("prompt export completed: file=%s promptCount=%d", filePath, len(names))
	return filePath, nil
}

func (a *App) PromptRootDir() (string, error) {
	_, root, err := a.promptService()
	if err != nil {
		return "", fmt.Errorf("load prompt root failed: %w", err)
	}
	return root, nil
}

func promptImportNames(items []promptcatalogapp.ImportPrompt) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}
