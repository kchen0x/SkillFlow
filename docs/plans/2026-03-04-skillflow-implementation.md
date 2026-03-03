# SkillFlow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build SkillFlow, a cross-platform desktop app (macOS + Windows) for managing LLM SKILLS across multiple tools, with GitHub install, cloud backup, and cross-tool sync.

**Architecture:** Go 1.26 core library (no UI deps) with interface-based extensibility for cloud providers/tool adapters/installers. Wails v2 bridges core to a React+TypeScript frontend via method bindings and channel-based event forwarding.

**Tech Stack:** Go 1.26, Wails v2, React 18, TypeScript, Zustand, Tailwind CSS, testify, zalando/go-keyring

---

## Phase 1: Foundation

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `wails.json`
- Create: `app/wails/main.go`
- Create: `frontend/package.json`

**Step 1: Initialize Go module**

```bash
cd /Users/shinerio/Workspace/code/SkillFlow
go mod init github.com/shinerio/skillflow
```

**Step 2: Install Wails CLI and init project**

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails init -n SkillFlow -t react-ts -d .
```

**Step 3: Create core directory structure**

```bash
mkdir -p core/{skill,sync,backup,install,update,config,registry,notify}
mkdir -p app/wails
mkdir -p docs/plans
```

**Step 4: Add Go dependencies**

```bash
go get github.com/stretchr/testify@latest
go get github.com/zalando/go-keyring@latest
go get github.com/google/uuid@latest
go get github.com/aliyun/aliyun-oss-go-sdk/v3@latest
go get github.com/tencentyun/cos-go-sdk-v5@latest
go get github.com/huaweicloud/huaweicloud-sdk-go-obs@latest
```

**Step 5: Verify Wails project runs**

```bash
wails dev
```
Expected: Browser opens with default Wails React template.

**Step 6: Commit**

```bash
git init
git add .
git commit -m "chore: initialize SkillFlow project with Wails + Go 1.26"
```

---

### Task 2: Core Data Models

**Files:**
- Create: `core/skill/model.go`
- Create: `core/config/model.go`
- Create: `core/notify/model.go`
- Create: `core/skill/model_test.go`

**Step 1: Write model tests**

Create `core/skill/model_test.go`:

```go
package skill_test

import (
    "testing"
    "time"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
)

func TestSkillSourceTypes(t *testing.T) {
    s := skill.Skill{
        ID:       "test-id",
        Name:     "my-skill",
        Source:   skill.SourceGitHub,
        Category: "coding",
    }
    assert.Equal(t, skill.SourceType("github"), s.Source)
    assert.True(t, s.IsGitHub())
    assert.False(t, s.IsManual())
}

func TestSkillIsManual(t *testing.T) {
    s := skill.Skill{Source: skill.SourceManual}
    assert.True(t, s.IsManual())
    assert.False(t, s.IsGitHub())
}

func TestSkillHasUpdate(t *testing.T) {
    s := skill.Skill{
        Source:       skill.SourceGitHub,
        SourceSHA:    "abc123",
        LatestSHA:    "def456",
    }
    assert.True(t, s.HasUpdate())

    s.LatestSHA = "abc123"
    assert.False(t, s.HasUpdate())
}
```

**Step 2: Run test to verify it fails**

```bash
cd core/skill && go test ./... -v
```
Expected: FAIL — package not defined.

**Step 3: Implement `core/skill/model.go`**

```go
package skill

import "time"

type SourceType string

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)

type Skill struct {
    ID            string
    Name          string
    Path          string
    Category      string
    Source        SourceType
    SourceURL     string
    SourceSubPath string
    SourceSHA     string
    LatestSHA     string
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}

func (s *Skill) IsGitHub() bool { return s.Source == SourceGitHub }
func (s *Skill) IsManual() bool { return s.Source == SourceManual }
func (s *Skill) HasUpdate() bool {
    return s.IsGitHub() && s.LatestSHA != "" && s.LatestSHA != s.SourceSHA
}
```

**Step 4: Implement `core/config/model.go`**

```go
package config

type ToolConfig struct {
    Name      string `json:"name"`
    SkillsDir string `json:"skillsDir"`
    Enabled   bool   `json:"enabled"`
    Custom    bool   `json:"custom"`
}

type CloudConfig struct {
    Provider    string            `json:"provider"`
    Enabled     bool              `json:"enabled"`
    BucketName  string            `json:"bucketName"`
    RemotePath  string            `json:"remotePath"`
    Credentials map[string]string `json:"credentials"`
}

type AppConfig struct {
    SkillsStorageDir string       `json:"skillsStorageDir"`
    DefaultCategory  string       `json:"defaultCategory"`
    Tools            []ToolConfig `json:"tools"`
    Cloud            CloudConfig  `json:"cloud"`
}
```

**Step 5: Implement `core/notify/model.go`**

```go
package notify

type EventType string

const (
    EventBackupStarted   EventType = "backup.started"
    EventBackupProgress  EventType = "backup.progress"
    EventBackupCompleted EventType = "backup.completed"
    EventBackupFailed    EventType = "backup.failed"
    EventSyncCompleted   EventType = "sync.completed"
    EventUpdateAvailable EventType = "update.available"
    EventSkillConflict   EventType = "skill.conflict"
)

type Event struct {
    Type    EventType `json:"type"`
    Payload any       `json:"payload"`
}

type BackupProgressPayload struct {
    FilesTotal    int    `json:"filesTotal"`
    FilesUploaded int    `json:"filesUploaded"`
    CurrentFile   string `json:"currentFile"`
}

type UpdateAvailablePayload struct {
    SkillID   string `json:"skillId"`
    SkillName string `json:"skillName"`
    CurrentSHA string `json:"currentSha"`
    LatestSHA  string `json:"latestSha"`
}

type ConflictPayload struct {
    SkillName  string `json:"skillName"`
    TargetPath string `json:"targetPath"`
}
```

**Step 6: Run tests**

```bash
go test ./core/... -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add core/
git commit -m "feat: add core data models (skill, config, notify)"
```

---

### Task 3: Notify Hub

**Files:**
- Create: `core/notify/hub.go`
- Create: `core/notify/hub_test.go`

**Step 1: Write failing test**

Create `core/notify/hub_test.go`:

```go
package notify_test

import (
    "testing"
    "time"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/stretchr/testify/assert"
)

func TestHubPublishSubscribe(t *testing.T) {
    hub := notify.NewHub()
    ch := hub.Subscribe()
    defer hub.Unsubscribe(ch)

    hub.Publish(notify.Event{Type: notify.EventBackupStarted, Payload: nil})

    select {
    case evt := <-ch:
        assert.Equal(t, notify.EventBackupStarted, evt.Type)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected event, got timeout")
    }
}

func TestHubMultipleSubscribers(t *testing.T) {
    hub := notify.NewHub()
    ch1 := hub.Subscribe()
    ch2 := hub.Subscribe()
    defer hub.Unsubscribe(ch1)
    defer hub.Unsubscribe(ch2)

    hub.Publish(notify.Event{Type: notify.EventSyncCompleted})

    for _, ch := range []<-chan notify.Event{ch1, ch2} {
        select {
        case evt := <-ch:
            assert.Equal(t, notify.EventSyncCompleted, evt.Type)
        case <-time.After(100 * time.Millisecond):
            t.Fatal("subscriber did not receive event")
        }
    }
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/notify/... -v
```
Expected: FAIL

**Step 3: Implement `core/notify/hub.go`**

```go
package notify

import "sync"

type Hub struct {
    mu          sync.RWMutex
    subscribers map[chan Event]struct{}
}

func NewHub() *Hub {
    return &Hub{subscribers: make(map[chan Event]struct{})}
}

func (h *Hub) Subscribe() <-chan Event {
    ch := make(chan Event, 32)
    h.mu.Lock()
    h.subscribers[ch] = struct{}{}
    h.mu.Unlock()
    return ch
}

func (h *Hub) Unsubscribe(ch <-chan Event) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for sub := range h.subscribers {
        if sub == ch {
            delete(h.subscribers, sub)
            close(sub)
            return
        }
    }
}

func (h *Hub) Publish(evt Event) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for sub := range h.subscribers {
        select {
        case sub <- evt:
        default: // drop if subscriber is slow
        }
    }
}
```

**Step 4: Run tests**

```bash
go test ./core/notify/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/notify/
git commit -m "feat: add channel-based notify hub"
```

---

### Task 4: Config Service

**Files:**
- Create: `core/config/service.go`
- Create: `core/config/service_test.go`
- Create: `core/config/defaults.go`

**Step 1: Write failing tests**

Create `core/config/service_test.go`:

```go
package config_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/config"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    cfg, err := svc.Load()
    require.NoError(t, err)
    assert.NotEmpty(t, cfg.SkillsStorageDir)
    assert.Equal(t, "Imported", cfg.DefaultCategory)
    assert.NotEmpty(t, cfg.Tools)
}

func TestSaveAndLoadConfig(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    cfg := config.DefaultConfig(dir)
    cfg.DefaultCategory = "MyCategory"
    err := svc.Save(cfg)
    require.NoError(t, err)

    loaded, err := svc.Load()
    require.NoError(t, err)
    assert.Equal(t, "MyCategory", loaded.DefaultCategory)
}

func TestConfigFileCreatedOnFirstLoad(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    _, err := svc.Load()
    require.NoError(t, err)
    _, err = os.Stat(filepath.Join(dir, "config.json"))
    assert.NoError(t, err)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/config/... -v
```
Expected: FAIL

**Step 3: Implement `core/config/defaults.go`**

```go
package config

import (
    "os"
    "path/filepath"
    "runtime"
)

func AppDataDir() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "SkillFlow")
    default: // darwin
        home, _ := os.UserHomeDir()
        return filepath.Join(home, "Library", "Application Support", "SkillFlow")
    }
}

func DefaultToolsDir(toolName string) string {
    home, _ := os.UserHomeDir()
    dirs := map[string]map[string]string{
        "darwin": {
            "claude-code": filepath.Join(home, ".claude", "skills"),
            "opencode":    filepath.Join(home, ".opencode", "skills"),
            "codex":       filepath.Join(home, ".codex", "skills"),
            "gemini-cli":  filepath.Join(home, ".gemini", "skills"),
            "openclaw":    filepath.Join(home, ".openclaw", "skills"),
        },
        "windows": {
            "claude-code": filepath.Join(os.Getenv("APPDATA"), "claude", "skills"),
            "opencode":    filepath.Join(os.Getenv("APPDATA"), "opencode", "skills"),
            "codex":       filepath.Join(os.Getenv("APPDATA"), "codex", "skills"),
            "gemini-cli":  filepath.Join(os.Getenv("APPDATA"), "gemini", "skills"),
            "openclaw":    filepath.Join(os.Getenv("APPDATA"), "openclaw", "skills"),
        },
    }
    goos := runtime.GOOS
    if goos != "windows" {
        goos = "darwin"
    }
    return dirs[goos][toolName]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
    tools := make([]ToolConfig, 0, len(builtinTools))
    for _, name := range builtinTools {
        dir := DefaultToolsDir(name)
        _, err := os.Stat(dir)
        tools = append(tools, ToolConfig{
            Name:      name,
            SkillsDir: dir,
            Enabled:   err == nil,
            Custom:    false,
        })
    }
    return AppConfig{
        SkillsStorageDir: filepath.Join(dataDir, "skills"),
        DefaultCategory:  "Imported",
        Tools:            tools,
        Cloud:            CloudConfig{RemotePath: "skillflow/"},
    }
}
```

**Step 4: Implement `core/config/service.go`**

```go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Service struct {
    dataDir    string
    configPath string
}

func NewService(dataDir string) *Service {
    return &Service{
        dataDir:    dataDir,
        configPath: filepath.Join(dataDir, "config.json"),
    }
}

func (s *Service) Load() (AppConfig, error) {
    if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
        cfg := DefaultConfig(s.dataDir)
        if err := s.Save(cfg); err != nil {
            return AppConfig{}, err
        }
        return cfg, nil
    }
    data, err := os.ReadFile(s.configPath)
    if err != nil {
        return AppConfig{}, err
    }
    var cfg AppConfig
    return cfg, json.Unmarshal(data, &cfg)
}

func (s *Service) Save(cfg AppConfig) error {
    if err := os.MkdirAll(s.dataDir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(s.configPath, data, 0644)
}
```

**Step 5: Run tests**

```bash
go test ./core/config/... -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add core/config/
git commit -m "feat: add config service with cross-platform defaults"
```

---

### Task 5: Registry

**Files:**
- Create: `core/registry/registry.go`

**Step 1: Implement registry (no tests needed — pure registration plumbing)**

```go
package registry

import (
    "github.com/shinerio/skillflow/core/backup"
    "github.com/shinerio/skillflow/core/install"
    "github.com/shinerio/skillflow/core/sync"
)

var (
    installers     = map[string]install.Installer{}
    adapters       = map[string]sync.ToolAdapter{}
    cloudProviders = map[string]backup.CloudProvider{}
)

func RegisterInstaller(i install.Installer)         { installers[i.Type()] = i }
func RegisterAdapter(a sync.ToolAdapter)             { adapters[a.Name()] = a }
func RegisterCloudProvider(p backup.CloudProvider)   { cloudProviders[p.Name()] = p }

func GetInstaller(t string) (install.Installer, bool) {
    i, ok := installers[t]
    return i, ok
}

func GetAdapter(name string) (sync.ToolAdapter, bool) {
    a, ok := adapters[name]
    return a, ok
}

func GetCloudProvider(name string) (backup.CloudProvider, bool) {
    p, ok := cloudProviders[name]
    return p, ok
}

func AllAdapters() []sync.ToolAdapter {
    result := make([]sync.ToolAdapter, 0, len(adapters))
    for _, a := range adapters {
        result = append(result, a)
    }
    return result
}

func AllCloudProviders() []backup.CloudProvider {
    result := make([]backup.CloudProvider, 0, len(cloudProviders))
    for _, p := range cloudProviders {
        result = append(result, p)
    }
    return result
}
```

**Step 2: Commit**

```bash
git add core/registry/
git commit -m "feat: add extensible registry for installers/adapters/providers"
```

---

## Phase 2: Core Skill Management

### Task 6: Skill Validator

**Files:**
- Create: `core/skill/validator.go`
- Create: `core/skill/validator_test.go`

**Step 1: Write failing tests**

```go
package skill_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidatorAcceptsDirectoryWithSKILLSmd(t *testing.T) {
    dir := t.TempDir()
    skillDir := filepath.Join(dir, "my-skill")
    require.NoError(t, os.MkdirAll(skillDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILLS.md"), []byte("# skill"), 0644))

    v := skill.NewValidator()
    err := v.Validate(skillDir)
    assert.NoError(t, err)
}

func TestValidatorRejectsDirectoryWithoutSKILLSmd(t *testing.T) {
    dir := t.TempDir()
    skillDir := filepath.Join(dir, "not-a-skill")
    require.NoError(t, os.MkdirAll(skillDir, 0755))

    v := skill.NewValidator()
    err := v.Validate(skillDir)
    assert.ErrorIs(t, err, skill.ErrNoSKILLSmd)
}

func TestValidatorRejectsNonDirectory(t *testing.T) {
    v := skill.NewValidator()
    err := v.Validate("/nonexistent/path")
    assert.Error(t, err)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/skill/... -run TestValidator -v
```

**Step 3: Implement `core/skill/validator.go`**

```go
package skill

import (
    "errors"
    "os"
    "path/filepath"
)

var ErrNoSKILLSmd = errors.New("SKILLS.md not found in skill directory")

// ValidationRule is the extension point for future complex validators.
type ValidationRule func(dir string) error

type Validator struct {
    rules []ValidationRule
}

func NewValidator(extraRules ...ValidationRule) *Validator {
    rules := []ValidationRule{requireSKILLSmd}
    return &Validator{rules: append(rules, extraRules...)}
}

func (v *Validator) Validate(dir string) error {
    for _, rule := range v.rules {
        if err := rule(dir); err != nil {
            return err
        }
    }
    return nil
}

func requireSKILLSmd(dir string) error {
    if _, err := os.Stat(dir); err != nil {
        return err
    }
    mdPath := filepath.Join(dir, "SKILLS.md")
    if _, err := os.Stat(mdPath); os.IsNotExist(err) {
        return ErrNoSKILLSmd
    }
    return nil
}
```

**Step 4: Run tests**

```bash
go test ./core/skill/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/skill/
git commit -m "feat: add extensible skill validator (SKILLS.md check)"
```

---

### Task 7: Skill Storage Service

**Files:**
- Create: `core/skill/storage.go`
- Create: `core/skill/storage_test.go`

**Step 1: Write failing tests**

```go
package skill_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func makeTestSkillDir(t *testing.T, baseDir, name string) string {
    t.Helper()
    dir := filepath.Join(baseDir, name)
    require.NoError(t, os.MkdirAll(dir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    return dir
}

func TestStorageListCategories(t *testing.T) {
    root := t.TempDir()
    svc := skill.NewStorage(root)
    require.NoError(t, svc.CreateCategory("coding"))
    require.NoError(t, svc.CreateCategory("writing"))
    cats, err := svc.ListCategories()
    require.NoError(t, err)
    assert.ElementsMatch(t, []string{"coding", "writing"}, cats)
}

func TestStorageImportSkill(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "my-skill")
    svc := skill.NewStorage(root)

    imported, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    require.NoError(t, err)
    assert.Equal(t, "my-skill", imported.Name)
    assert.Equal(t, "coding", imported.Category)

    // verify directory was copied
    _, err = os.Stat(filepath.Join(root, "coding", "my-skill", "SKILLS.md"))
    assert.NoError(t, err)
}

func TestStorageConflictDetected(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "dup-skill")
    svc := skill.NewStorage(root)

    _, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    require.NoError(t, err)

    _, err = svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    assert.ErrorIs(t, err, skill.ErrSkillExists)
}

func TestStorageDeleteSkill(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "del-skill")
    svc := skill.NewStorage(root)

    s, err := svc.Import(skillDir, "", skill.SourceManual, "", "")
    require.NoError(t, err)
    require.NoError(t, svc.Delete(s.ID))

    skills, err := svc.ListAll()
    require.NoError(t, err)
    assert.Empty(t, skills)
}

func TestStorageMoveCategory(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "move-skill")
    svc := skill.NewStorage(root)
    require.NoError(t, svc.CreateCategory("cat-a"))
    require.NoError(t, svc.CreateCategory("cat-b"))

    s, err := svc.Import(skillDir, "cat-a", skill.SourceManual, "", "")
    require.NoError(t, err)

    err = svc.MoveCategory(s.ID, "cat-b")
    require.NoError(t, err)

    updated, err := svc.Get(s.ID)
    require.NoError(t, err)
    assert.Equal(t, "cat-b", updated.Category)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/skill/... -run TestStorage -v
```

**Step 3: Implement `core/skill/storage.go`**

```go
package skill

import (
    "encoding/json"
    "errors"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/google/uuid"
)

var ErrSkillExists = errors.New("skill already exists in target location")
var ErrSkillNotFound = errors.New("skill not found")

type Storage struct {
    root    string
    metaDir string
}

func NewStorage(root string) *Storage {
    return &Storage{
        root:    root,
        metaDir: filepath.Join(filepath.Dir(root), "meta"),
    }
}

func (s *Storage) CreateCategory(name string) error {
    return os.MkdirAll(filepath.Join(s.root, name), 0755)
}

func (s *Storage) ListCategories() ([]string, error) {
    entries, err := os.ReadDir(s.root)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }
    var cats []string
    for _, e := range entries {
        if e.IsDir() {
            cats = append(cats, e.Name())
        }
    }
    return cats, nil
}

func (s *Storage) Import(srcDir, category string, source SourceType, sourceURL, sourceSubPath string) (*Skill, error) {
    name := filepath.Base(srcDir)
    targetDir := filepath.Join(s.root, category, name)
    if _, err := os.Stat(targetDir); err == nil {
        return nil, ErrSkillExists
    }
    if err := copyDir(srcDir, targetDir); err != nil {
        return nil, err
    }
    sk := &Skill{
        ID:            uuid.New().String(),
        Name:          name,
        Path:          targetDir,
        Category:      category,
        Source:        source,
        SourceURL:     sourceURL,
        SourceSubPath: sourceSubPath,
        InstalledAt:   time.Now(),
        UpdatedAt:     time.Now(),
    }
    return sk, s.saveMeta(sk)
}

func (s *Storage) Get(id string) (*Skill, error) {
    skills, err := s.ListAll()
    if err != nil {
        return nil, err
    }
    for _, sk := range skills {
        if sk.ID == id {
            return sk, nil
        }
    }
    return nil, ErrSkillNotFound
}

func (s *Storage) ListAll() ([]*Skill, error) {
    if err := os.MkdirAll(s.metaDir, 0755); err != nil {
        return nil, err
    }
    entries, err := os.ReadDir(s.metaDir)
    if err != nil {
        return nil, err
    }
    var skills []*Skill
    for _, e := range entries {
        if filepath.Ext(e.Name()) != ".json" {
            continue
        }
        data, err := os.ReadFile(filepath.Join(s.metaDir, e.Name()))
        if err != nil {
            continue
        }
        var sk Skill
        if err := json.Unmarshal(data, &sk); err == nil {
            skills = append(skills, &sk)
        }
    }
    return skills, nil
}

func (s *Storage) Delete(id string) error {
    sk, err := s.Get(id)
    if err != nil {
        return err
    }
    if err := os.RemoveAll(sk.Path); err != nil {
        return err
    }
    return os.Remove(filepath.Join(s.metaDir, id+".json"))
}

func (s *Storage) MoveCategory(id, newCategory string) error {
    sk, err := s.Get(id)
    if err != nil {
        return err
    }
    newPath := filepath.Join(s.root, newCategory, sk.Name)
    if err := os.MkdirAll(filepath.Join(s.root, newCategory), 0755); err != nil {
        return err
    }
    if err := os.Rename(sk.Path, newPath); err != nil {
        return err
    }
    sk.Path = newPath
    sk.Category = newCategory
    sk.UpdatedAt = time.Now()
    return s.saveMeta(sk)
}

func (s *Storage) UpdateMeta(sk *Skill) error {
    sk.UpdatedAt = time.Now()
    return s.saveMeta(sk)
}

func (s *Storage) RenameCategory(oldName, newName string) error {
    oldPath := filepath.Join(s.root, oldName)
    newPath := filepath.Join(s.root, newName)
    if err := os.Rename(oldPath, newPath); err != nil {
        return err
    }
    // Update all skill metadata in this category
    skills, err := s.ListAll()
    if err != nil {
        return err
    }
    for _, sk := range skills {
        if sk.Category == oldName {
            sk.Category = newName
            sk.Path = filepath.Join(newPath, sk.Name)
            sk.UpdatedAt = time.Now()
            if err := s.saveMeta(sk); err != nil {
                return err
            }
        }
    }
    return nil
}

func (s *Storage) DeleteCategory(name string) error {
    skills, err := s.ListAll()
    if err != nil {
        return err
    }
    // Move skills to uncategorized before deleting
    for _, sk := range skills {
        if sk.Category == name {
            if err := s.MoveCategory(sk.ID, ""); err != nil {
                return err
            }
        }
    }
    return os.Remove(filepath.Join(s.root, name))
}

// OverwriteFromDir replaces an existing skill's directory contents from srcDir, used for updates.
func (s *Storage) OverwriteFromDir(id, srcDir string) error {
    sk, err := s.Get(id)
    if err != nil {
        return err
    }
    if err := os.RemoveAll(sk.Path); err != nil {
        return err
    }
    return copyDir(srcDir, sk.Path)
}

func (s *Storage) saveMeta(sk *Skill) error {
    if err := os.MkdirAll(s.metaDir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(sk, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(s.metaDir, sk.ID+".json"), data, 0644)
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(src, path)
        target := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(target, info.Mode())
        }
        return copyFile(path, target)
    })
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}
```

**Step 4: Run tests**

```bash
go test ./core/skill/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/skill/
git commit -m "feat: add skill storage service with category and meta management"
```

---

## Phase 3: Install

### Task 8: Install Interfaces

**Files:**
- Create: `core/install/installer.go`

```go
package install

import "context"

type InstallSource struct {
    Type string // "github" | "local"
    URI  string
}

type SkillCandidate struct {
    Name      string
    Path      string // relative path within source
    Installed bool
}

type Installer interface {
    Type() string
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
```

**Step 1: Commit interface**

```bash
git add core/install/
git commit -m "feat: add installer interface"
```

---

### Task 9: GitHub Installer

**Files:**
- Create: `core/install/github.go`
- Create: `core/install/github_test.go`

**Step 1: Write failing tests (using httptest to mock GitHub API)**

```go
package install_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/shinerio/skillflow/core/install"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func mockGitHubServer(t *testing.T) *httptest.Server {
    t.Helper()
    mux := http.NewServeMux()
    // Mock: list skills directory contents
    mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
        items := []map[string]any{
            {"name": "skill-a", "type": "dir", "path": "skills/skill-a"},
            {"name": "skill-b", "type": "dir", "path": "skills/skill-b"},
            {"name": "readme.md", "type": "file", "path": "skills/readme.md"},
        }
        json.NewEncoder(w).Encode(items)
    })
    // Mock: check SKILLS.md existence for skill-a (returns file info)
    mux.HandleFunc("/repos/user/repo/contents/skills/skill-a/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]any{"name": "SKILLS.md", "type": "file"})
    })
    // Mock: skill-b has no SKILLS.md (404)
    mux.HandleFunc("/repos/user/repo/contents/skills/skill-b/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    })
    return httptest.NewServer(mux)
}

func TestGitHubInstallerScan(t *testing.T) {
    srv := mockGitHubServer(t)
    defer srv.Close()

    installer := install.NewGitHubInstaller(srv.URL)
    candidates, err := installer.Scan(context.Background(), install.InstallSource{
        Type: "github",
        URI:  srv.URL + "/repos/user/repo",
    })
    require.NoError(t, err)
    // Only skill-a has SKILLS.md, skill-b does not
    assert.Len(t, candidates, 1)
    assert.Equal(t, "skill-a", candidates[0].Name)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/install/... -v
```

**Step 3: Implement `core/install/github.go`**

```go
package install

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type githubContent struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Path        string `json:"path"`
    DownloadURL string `json:"download_url"`
}

type GitHubInstaller struct {
    baseURL string // overridable for tests
    client  *http.Client
}

func NewGitHubInstaller(baseURL string) *GitHubInstaller {
    if baseURL == "" {
        baseURL = "https://api.github.com"
    }
    return &GitHubInstaller{baseURL: baseURL, client: http.DefaultClient}
}

func (g *GitHubInstaller) Type() string { return "github" }

func (g *GitHubInstaller) Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error) {
    owner, repo, err := parseGitHubURI(source.URI)
    if err != nil {
        return nil, err
    }
    items, err := g.listContents(ctx, owner, repo, "skills")
    if err != nil {
        return nil, err
    }
    var candidates []SkillCandidate
    for _, item := range items {
        if item.Type != "dir" {
            continue
        }
        // Check SKILLS.md exists
        if g.fileExists(ctx, owner, repo, item.Path+"/SKILLS.md") {
            candidates = append(candidates, SkillCandidate{
                Name: item.Name,
                Path: item.Path,
            })
        }
    }
    return candidates, nil
}

func (g *GitHubInstaller) Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error {
    owner, repo, err := parseGitHubURI(source.URI)
    if err != nil {
        return err
    }
    for _, c := range selected {
        if err := g.downloadDir(ctx, owner, repo, c.Path, category, c.Name); err != nil {
            return fmt.Errorf("install %s: %w", c.Name, err)
        }
    }
    return nil
}

func (g *GitHubInstaller) listContents(ctx context.Context, owner, repo, path string) ([]githubContent, error) {
    url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, path)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var items []githubContent
    return items, json.NewDecoder(resp.Body).Decode(&items)
}

func (g *GitHubInstaller) fileExists(ctx context.Context, owner, repo, path string) bool {
    url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, path)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return false
    }
    resp.Body.Close()
    return resp.StatusCode == http.StatusOK
}

func (g *GitHubInstaller) downloadDir(ctx context.Context, owner, repo, remotePath, category, name string) error {
    items, err := g.listContents(ctx, owner, repo, remotePath)
    if err != nil {
        return err
    }
    for _, item := range items {
        if item.Type == "dir" {
            if err := g.downloadDir(ctx, owner, repo, item.Path, category, name); err != nil {
                return err
            }
        } else if item.DownloadURL != "" {
            if err := g.downloadFile(ctx, item.DownloadURL, category, name, item.Path, remotePath); err != nil {
                return err
            }
        }
    }
    return nil
}

func (g *GitHubInstaller) downloadFile(ctx context.Context, url, category, skillName, filePath, basePath string) error {
    rel := strings.TrimPrefix(filePath, basePath+"/")
    // caller (app layer) sets actual target; installer returns to tmp dir
    tmpDir := filepath.Join(os.TempDir(), "skillflow-install", skillName)
    target := filepath.Join(tmpDir, rel)
    if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
        return err
    }
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    f, err := os.Create(target)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, resp.Body)
    return err
}

// DownloadTo downloads a skill candidate from GitHub into targetDir.
// Called by app layer after scanning, before importing into storage.
func (g *GitHubInstaller) DownloadTo(ctx context.Context, source InstallSource, c SkillCandidate, targetDir string) error {
    owner, repo, err := parseGitHubURI(source.URI)
    if err != nil {
        return err
    }
    return g.downloadDir(ctx, owner, repo, c.Path, "", c.Name)
}

// GetLatestSHA fetches the latest commit SHA for a skill's subdirectory path.
func (g *GitHubInstaller) GetLatestSHA(ctx context.Context, repoURL, subPath string) (string, error) {
    owner, repo, err := parseGitHubURI(repoURL)
    if err != nil {
        return "", err
    }
    url := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", g.baseURL, owner, repo, subPath)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    var commits []struct{ SHA string `json:"sha"` }
    if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
        return "", err
    }
    return commits[0].SHA, nil
}

func parseGitHubURI(uri string) (owner, repo string, err error) {
    // Accept: https://github.com/owner/repo or https://api.github.com/repos/owner/repo
    uri = strings.TrimSuffix(uri, "/")
    parts := strings.Split(uri, "/")
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid GitHub URI: %s", uri)
    }
    return parts[len(parts)-2], parts[len(parts)-1], nil
}
```

**Step 4: Run tests**

```bash
go test ./core/install/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/install/
git commit -m "feat: add GitHub installer with SKILLS.md validation"
```

---

### Task 10: Local Installer

**Files:**
- Create: `core/install/local.go`
- Create: `core/install/local_test.go`

**Step 1: Write failing tests**

```go
package install_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/install"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLocalInstallerScanValidSkill(t *testing.T) {
    dir := t.TempDir()
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# skill"), 0644))

    inst := install.NewLocalInstaller()
    candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
    require.NoError(t, err)
    assert.Len(t, candidates, 1)
    assert.Equal(t, filepath.Base(dir), candidates[0].Name)
}

func TestLocalInstallerScanInvalidSkill(t *testing.T) {
    dir := t.TempDir() // no SKILLS.md
    inst := install.NewLocalInstaller()
    candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
    require.NoError(t, err)
    assert.Empty(t, candidates)
}
```

**Step 2: Implement `core/install/local.go`**

```go
package install

import (
    "context"
    "os"
    "path/filepath"
    "github.com/shinerio/skillflow/core/skill"
)

type LocalInstaller struct {
    validator *skill.Validator
}

func NewLocalInstaller() *LocalInstaller {
    return &LocalInstaller{validator: skill.NewValidator()}
}

func (l *LocalInstaller) Type() string { return "local" }

func (l *LocalInstaller) Scan(_ context.Context, source InstallSource) ([]SkillCandidate, error) {
    dir := source.URI
    if err := l.validator.Validate(dir); err != nil {
        return nil, nil // not a valid skill dir — return empty, not error
    }
    return []SkillCandidate{{Name: filepath.Base(dir), Path: dir}}, nil
}

func (l *LocalInstaller) Install(_ context.Context, _ InstallSource, selected []SkillCandidate, _ string) error {
    // Local install: the app layer copies from candidate.Path directly via Storage.Import
    // This installer's Install is a no-op; the app layer calls Storage.Import
    _ = selected
    return nil
}

// Ensure os import used
var _ = os.Stat
```

**Step 3: Run tests**

```bash
go test ./core/install/... -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/install/
git commit -m "feat: add local installer for manual skill import"
```

---

## Phase 4: Sync

### Task 11: Sync Interfaces and Tool Adapters

**Files:**
- Create: `core/sync/adapter.go`
- Create: `core/sync/filesystem_adapter.go`
- Create: `core/sync/filesystem_adapter_test.go`

**Step 1: Define sync interface**

```go
// core/sync/adapter.go
package sync

import (
    "context"
    "github.com/shinerio/skillflow/core/skill"
)

type ToolAdapter interface {
    Name() string
    DefaultSkillsDir() string
    // Push copies skills into targetDir, flattened (no category subdirs)
    Push(ctx context.Context, skills []*skill.Skill, targetDir string) error
    // Pull scans sourceDir and returns skill candidates (not yet imported)
    Pull(ctx context.Context, sourceDir string) ([]*skill.Skill, error)
}
```

**Step 2: Write failing tests**

```go
// core/sync/filesystem_adapter_test.go
package sync_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    toolsync "github.com/shinerio/skillflow/core/sync"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func makeSkillDir(t *testing.T, root, category, name string) *skill.Skill {
    t.Helper()
    dir := filepath.Join(root, category, name)
    require.NoError(t, os.MkdirAll(dir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    return &skill.Skill{Name: name, Path: dir, Category: category}
}

func TestFilesystemAdapterPushFlattens(t *testing.T) {
    src := t.TempDir()
    dst := t.TempDir()
    sk := makeSkillDir(t, src, "coding", "my-skill")

    adapter := toolsync.NewFilesystemAdapter("test-tool", "")
    err := adapter.Push(context.Background(), []*skill.Skill{sk}, dst)
    require.NoError(t, err)

    // skill should be at dst/my-skill (no category subdir)
    _, err = os.Stat(filepath.Join(dst, "my-skill", "SKILLS.md"))
    assert.NoError(t, err)
}

func TestFilesystemAdapterPull(t *testing.T) {
    src := t.TempDir()
    // Create two valid skills directly in src (tool dir is flat)
    for _, name := range []string{"skill-x", "skill-y"} {
        dir := filepath.Join(src, name)
        require.NoError(t, os.MkdirAll(dir, 0755))
        require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    }
    // Create a non-skill directory
    require.NoError(t, os.MkdirAll(filepath.Join(src, "not-a-skill"), 0755))

    adapter := toolsync.NewFilesystemAdapter("test-tool", "")
    skills, err := adapter.Pull(context.Background(), src)
    require.NoError(t, err)
    assert.Len(t, skills, 2)
}
```

**Step 3: Implement `core/sync/filesystem_adapter.go`**

```go
package sync

import (
    "context"
    "io"
    "os"
    "path/filepath"
    "github.com/shinerio/skillflow/core/skill"
)

// FilesystemAdapter works for all tools — they all share the same file-based skills directory model.
type FilesystemAdapter struct {
    name          string
    defaultSkillsDir string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
    return &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skill.Skill, targetDir string) error {
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        return err
    }
    for _, sk := range skills {
        dst := filepath.Join(targetDir, sk.Name)
        if err := copyDir(sk.Path, dst); err != nil {
            return err
        }
    }
    return nil
}

func (f *FilesystemAdapter) Pull(_ context.Context, sourceDir string) ([]*skill.Skill, error) {
    validator := skill.NewValidator()
    entries, err := os.ReadDir(sourceDir)
    if err != nil {
        return nil, err
    }
    var skills []*skill.Skill
    for _, e := range entries {
        if !e.IsDir() {
            continue
        }
        dir := filepath.Join(sourceDir, e.Name())
        if err := validator.Validate(dir); err == nil {
            skills = append(skills, &skill.Skill{
                Name:   e.Name(),
                Path:   dir,
                Source: skill.SourceManual, // pulled from external tool
            })
        }
    }
    return skills, nil
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(src, path)
        target := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(target, info.Mode())
        }
        return copyFile(path, target)
    })
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}
```

**Step 4: Run tests**

```bash
go test ./core/sync/... -v
```
Expected: PASS

**Step 5: Register all built-in adapters in `app/wails/adapters.go`**

```go
package main

import (
    "github.com/shinerio/skillflow/core/config"
    "github.com/shinerio/skillflow/core/registry"
    toolsync "github.com/shinerio/skillflow/core/sync"
    "runtime"
)

func registerAdapters() {
    tools := []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}
    for _, name := range tools {
        registry.RegisterAdapter(toolsync.NewFilesystemAdapter(name, config.DefaultToolsDir(name)))
    }
}
```

**Step 6: Commit**

```bash
git add core/sync/ app/wails/adapters.go
git commit -m "feat: add filesystem sync adapter shared by all tools"
```

---

## Phase 5: Cloud Backup

### Task 12: Cloud Backup Interface + Aliyun OSS

**Files:**
- Create: `core/backup/provider.go`
- Create: `core/backup/aliyun.go`

**Step 1: Define backup interface**

```go
// core/backup/provider.go
package backup

import "context"

type CredentialField struct {
    Key         string `json:"key"`
    Label       string `json:"label"`
    Placeholder string `json:"placeholder"`
    Secret      bool   `json:"secret"`
}

type RemoteFile struct {
    Path         string `json:"path"`
    Size         int64  `json:"size"`
    IsDir        bool   `json:"isDir"`
}

type CloudProvider interface {
    Name() string
    Init(credentials map[string]string) error
    // Sync mirrors localDir to cloud bucket at remotePath (incremental, no compression)
    Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
    Restore(ctx context.Context, bucket, remotePath, localDir string) error
    List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
    RequiredCredentials() []CredentialField
}
```

**Step 2: Implement `core/backup/aliyun.go`**

```go
package backup

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "github.com/aliyun/aliyun-oss-go-sdk/v3/oss"
)

type AliyunProvider struct {
    client *oss.Client
}

func NewAliyunProvider() *AliyunProvider { return &AliyunProvider{} }

func (a *AliyunProvider) Name() string { return "aliyun" }

func (a *AliyunProvider) RequiredCredentials() []CredentialField {
    return []CredentialField{
        {Key: "access_key_id", Label: "Access Key ID", Secret: false},
        {Key: "access_key_secret", Label: "Access Key Secret", Secret: true},
        {Key: "endpoint", Label: "Endpoint", Placeholder: "oss-cn-hangzhou.aliyuncs.com"},
    }
}

func (a *AliyunProvider) Init(creds map[string]string) error {
    client, err := oss.New(creds["endpoint"], creds["access_key_id"], creds["access_key_secret"])
    if err != nil {
        return err
    }
    a.client = client
    return nil
}

func (a *AliyunProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return err
    }
    return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }
        rel, _ := filepath.Rel(localDir, path)
        key := remotePath + strings.ReplaceAll(rel, string(filepath.Separator), "/")
        if onProgress != nil {
            onProgress(rel)
        }
        return b.PutObjectFromFile(key, path)
    })
}

func (a *AliyunProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return err
    }
    marker := ""
    for {
        result, err := b.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
        if err != nil {
            return err
        }
        for _, obj := range result.Objects {
            rel := strings.TrimPrefix(obj.Key, remotePath)
            local := filepath.Join(localDir, filepath.FromSlash(rel))
            if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
                return err
            }
            if err := b.GetObjectToFile(obj.Key, local); err != nil {
                return err
            }
        }
        if !result.IsTruncated {
            break
        }
        marker = result.NextMarker
    }
    return nil
}

func (a *AliyunProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return nil, err
    }
    result, err := b.ListObjects(oss.Prefix(remotePath))
    if err != nil {
        return nil, err
    }
    var files []RemoteFile
    for _, obj := range result.Objects {
        files = append(files, RemoteFile{
            Path: strings.TrimPrefix(obj.Key, remotePath),
            Size: obj.Size,
        })
    }
    return files, nil
}
```

**Step 3: Implement Tencent COS and Huawei OBS similarly in `core/backup/tencent.go` and `core/backup/huawei.go`** (same interface, different SDK calls — follow same pattern as AliyunProvider)

**Step 4: Register providers in `app/wails/providers.go`**

```go
package main

import (
    "github.com/shinerio/skillflow/core/backup"
    "github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
    registry.RegisterCloudProvider(backup.NewAliyunProvider())
    registry.RegisterCloudProvider(backup.NewTencentProvider())
    registry.RegisterCloudProvider(backup.NewHuaweiProvider())
}
```

**Step 5: Commit**

```bash
git add core/backup/ app/wails/providers.go
git commit -m "feat: add cloud backup interface + Aliyun/Tencent/Huawei providers"
```

---

## Phase 6: Update Checker

### Task 13: GitHub Update Checker

**Files:**
- Create: `core/update/checker.go`
- Create: `core/update/checker_test.go`

**Step 1: Write failing tests (using httptest)**

```go
package update_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/shinerio/skillflow/core/update"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCheckerDetectsUpdate(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode([]map[string]any{{"sha": "newsha123"}})
    }))
    defer srv.Close()

    checker := update.NewChecker(srv.URL)
    sk := &skill.Skill{
        Source:        skill.SourceGitHub,
        SourceURL:     "https://github.com/user/repo",
        SourceSubPath: "skills/skill-a",
        SourceSHA:     "oldsha456",
    }
    result, err := checker.Check(context.Background(), sk)
    require.NoError(t, err)
    assert.True(t, result.HasUpdate)
    assert.Equal(t, "newsha123", result.LatestSHA)
}

func TestCheckerNoUpdateWhenSHAMatches(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode([]map[string]any{{"sha": "sameSHA"}})
    }))
    defer srv.Close()

    checker := update.NewChecker(srv.URL)
    sk := &skill.Skill{
        Source:    skill.SourceGitHub,
        SourceSHA: "sameSHA",
    }
    result, err := checker.Check(context.Background(), sk)
    require.NoError(t, err)
    assert.False(t, result.HasUpdate)
}
```

**Step 2: Implement `core/update/checker.go`**

```go
package update

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "github.com/shinerio/skillflow/core/skill"
)

type CheckResult struct {
    SkillID   string
    HasUpdate bool
    LatestSHA string
}

type Checker struct {
    baseURL string
    client  *http.Client
}

func NewChecker(baseURL string) *Checker {
    if baseURL == "" {
        baseURL = "https://api.github.com"
    }
    return &Checker{baseURL: baseURL, client: http.DefaultClient}
}

func (c *Checker) Check(ctx context.Context, sk *skill.Skill) (CheckResult, error) {
    if !sk.IsGitHub() {
        return CheckResult{}, nil
    }
    owner, repo, subPath := parseSourceURL(sk.SourceURL, sk.SourceSubPath)
    url := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", c.baseURL, owner, repo, subPath)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := c.client.Do(req)
    if err != nil {
        return CheckResult{}, err
    }
    defer resp.Body.Close()

    var commits []struct{ SHA string `json:"sha"` }
    if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
        return CheckResult{}, err
    }
    latestSHA := commits[0].SHA
    return CheckResult{
        SkillID:   sk.ID,
        LatestSHA: latestSHA,
        HasUpdate: latestSHA != sk.SourceSHA,
    }, nil
}

func parseSourceURL(sourceURL, subPath string) (owner, repo, path string) {
    sourceURL = strings.TrimSuffix(sourceURL, "/")
    parts := strings.Split(sourceURL, "/")
    owner = parts[len(parts)-2]
    repo = parts[len(parts)-1]
    return owner, repo, subPath
}
```

**Step 3: Run tests**

```bash
go test ./core/update/... -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/update/
git commit -m "feat: add GitHub SHA-based update checker"
```

---

## Phase 7: Wails App Layer

### Task 14: Wails App Methods

**Files:**
- Modify: `app/wails/app.go`
- Create: `app/wails/events.go`

**Step 1: Implement `app/wails/app.go`** (all methods exposed to frontend)

```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "github.com/shinerio/skillflow/core/backup"
    "github.com/shinerio/skillflow/core/config"
    "github.com/shinerio/skillflow/core/install"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/shinerio/skillflow/core/registry"
    "github.com/shinerio/skillflow/core/skill"
    toolsync "github.com/shinerio/skillflow/core/sync"
    "github.com/shinerio/skillflow/core/update"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
    ctx     context.Context
    hub     *notify.Hub
    storage *skill.Storage
    config  *config.Service
    checker *update.Checker
}

func NewApp() *App {
    return &App{hub: notify.NewHub()}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    dataDir := config.AppDataDir()
    a.config = config.NewService(dataDir)
    cfg, _ := a.config.Load()
    a.storage = skill.NewStorage(cfg.SkillsStorageDir)
    a.checker = update.NewChecker("")
    registerAdapters()
    registerProviders()
    go forwardEvents(ctx, a.hub)
    go a.checkUpdatesOnStartup()
}

// autoBackup triggers cloud backup after any mutating operation if cloud is enabled.
func (a *App) autoBackup() {
    cfg, err := a.config.Load()
    if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider == "" {
        return
    }
    provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
    if !ok {
        return
    }
    if err := provider.Init(cfg.Cloud.Credentials); err != nil {
        return
    }
    a.hub.Publish(notify.Event{Type: notify.EventBackupStarted})
    err = provider.Sync(a.ctx, cfg.SkillsStorageDir, cfg.Cloud.BucketName, cfg.Cloud.RemotePath,
        func(file string) {
            a.hub.Publish(notify.Event{
                Type:    notify.EventBackupProgress,
                Payload: notify.BackupProgressPayload{CurrentFile: file},
            })
        })
    if err != nil {
        a.hub.Publish(notify.Event{Type: notify.EventBackupFailed, Payload: err.Error()})
    } else {
        a.hub.Publish(notify.Event{Type: notify.EventBackupCompleted})
    }
}

// --- Skills ---

func (a *App) ListSkills() ([]*skill.Skill, error) {
    return a.storage.ListAll()
}

func (a *App) ListCategories() ([]string, error) {
    return a.storage.ListCategories()
}

func (a *App) CreateCategory(name string) error {
    return a.storage.CreateCategory(name)
}

func (a *App) RenameCategory(oldName, newName string) error {
    return a.storage.RenameCategory(oldName, newName)
}

func (a *App) DeleteCategory(name string) error {
    return a.storage.DeleteCategory(name)
}

func (a *App) MoveSkillCategory(skillID, category string) error {
    return a.storage.MoveCategory(skillID, category)
}

func (a *App) DeleteSkill(skillID string) error {
    if err := a.storage.Delete(skillID); err != nil {
        return err
    }
    go a.autoBackup()
    return nil
}

// --- Install ---

// ScanGitHub scans a GitHub repo for valid skills, marking already-installed ones.
func (a *App) ScanGitHub(repoURL string) ([]install.SkillCandidate, error) {
    inst := install.NewGitHubInstaller("")
    candidates, err := inst.Scan(a.ctx, install.InstallSource{Type: "github", URI: repoURL})
    if err != nil {
        return nil, err
    }
    // Mark already-installed skills
    existing, _ := a.storage.ListAll()
    existingNames := map[string]bool{}
    for _, sk := range existing {
        existingNames[sk.Name] = true
    }
    for i := range candidates {
        candidates[i].Installed = existingNames[candidates[i].Name]
    }
    return candidates, nil
}

// InstallFromGitHub downloads selected skills from GitHub and imports them into storage.
func (a *App) InstallFromGitHub(repoURL string, candidates []install.SkillCandidate, category string) error {
    inst := install.NewGitHubInstaller("")
    source := install.InstallSource{Type: "github", URI: repoURL}

    // Download each candidate to tmp, then import via storage
    for _, c := range candidates {
        tmpDir := filepath.Join(os.TempDir(), "skillflow-install", c.Name)
        defer os.RemoveAll(tmpDir)

        if err := inst.DownloadTo(a.ctx, source, c, tmpDir); err != nil {
            return fmt.Errorf("download %s: %w", c.Name, err)
        }

        // Get commit SHA for the skill directory
        sha, _ := inst.GetLatestSHA(a.ctx, repoURL, c.Path)

        _, err := a.storage.Import(tmpDir, category, skill.SourceGitHub, repoURL, c.Path)
        if err != nil {
            return err
        }
        // Update SHA in meta after import
        skills, _ := a.storage.ListAll()
        for _, sk := range skills {
            if sk.Name == c.Name && sk.SourceURL == repoURL {
                sk.SourceSHA = sha
                _ = a.storage.UpdateMeta(sk)
                break
            }
        }
    }
    go a.autoBackup()
    return nil
}

func (a *App) ImportLocal(dir, category string) (*skill.Skill, error) {
    sk, err := a.storage.Import(dir, category, skill.SourceManual, "", "")
    if err != nil {
        return nil, err
    }
    go a.autoBackup()
    return sk, nil
}

// --- Sync ---

func (a *App) GetEnabledTools() ([]config.ToolConfig, error) {
    cfg, err := a.config.Load()
    if err != nil {
        return nil, err
    }
    var enabled []config.ToolConfig
    for _, t := range cfg.Tools {
        if t.Enabled {
            enabled = append(enabled, t)
        }
    }
    return enabled, nil
}

// ScanToolSkills lists all skills in a tool's directory for the pull page.
func (a *App) ScanToolSkills(toolName string) ([]*skill.Skill, error) {
    cfg, _ := a.config.Load()
    for _, t := range cfg.Tools {
        if t.Name == toolName {
            adapter := getAdapter(t)
            return adapter.Pull(a.ctx, t.SkillsDir)
        }
    }
    return nil, nil
}

// PushToTools pushes selected skills to target tools, flattened (no category dirs).
// Returns list of conflict skill names that were skipped.
func (a *App) PushToTools(skillIDs []string, toolNames []string) ([]string, error) {
    cfg, _ := a.config.Load()
    skills, err := a.storage.ListAll()
    if err != nil {
        return nil, err
    }

    // Filter selected skills
    idSet := map[string]bool{}
    for _, id := range skillIDs {
        idSet[id] = true
    }
    var selected []*skill.Skill
    for _, sk := range skills {
        if idSet[sk.ID] {
            selected = append(selected, sk)
        }
    }

    var conflicts []string
    for _, toolName := range toolNames {
        for _, t := range cfg.Tools {
            if t.Name != toolName {
                continue
            }
            adapter := getAdapter(t)
            for _, sk := range selected {
                dst := filepath.Join(t.SkillsDir, sk.Name)
                if _, err := os.Stat(dst); err == nil {
                    conflicts = append(conflicts, fmt.Sprintf("%s -> %s", sk.Name, toolName))
                    continue
                }
            }
            _ = adapter.Push(a.ctx, selected, t.SkillsDir)
        }
    }
    return conflicts, nil
}

// PushToToolsForce pushes and overwrites conflicts.
func (a *App) PushToToolsForce(skillIDs []string, toolNames []string) error {
    cfg, _ := a.config.Load()
    skills, _ := a.storage.ListAll()
    idSet := map[string]bool{}
    for _, id := range skillIDs {
        idSet[id] = true
    }
    var selected []*skill.Skill
    for _, sk := range skills {
        if idSet[sk.ID] {
            selected = append(selected, sk)
        }
    }
    for _, toolName := range toolNames {
        for _, t := range cfg.Tools {
            if t.Name == toolName {
                _ = getAdapter(t).Push(a.ctx, selected, t.SkillsDir)
            }
        }
    }
    return nil
}

// PullFromTool imports selected skills from a tool into SkillFlow storage.
func (a *App) PullFromTool(toolName string, skillNames []string, category string) ([]string, error) {
    cfg, _ := a.config.Load()
    nameSet := map[string]bool{}
    for _, n := range skillNames {
        nameSet[n] = true
    }
    for _, t := range cfg.Tools {
        if t.Name != toolName {
            continue
        }
        adapter := getAdapter(t)
        candidates, err := adapter.Pull(a.ctx, t.SkillsDir)
        if err != nil {
            return nil, err
        }
        var conflicts []string
        for _, sk := range candidates {
            if !nameSet[sk.Name] {
                continue
            }
            _, err := a.storage.Import(sk.Path, category, skill.SourceManual, "", "")
            if err != nil {
                if err == skill.ErrSkillExists {
                    conflicts = append(conflicts, sk.Name)
                }
            }
        }
        go a.autoBackup()
        return conflicts, nil
    }
    return nil, nil
}

// PullFromToolForce imports selected skills, overwriting existing ones.
func (a *App) PullFromToolForce(toolName string, skillNames []string, category string) error {
    cfg, _ := a.config.Load()
    nameSet := map[string]bool{}
    for _, n := range skillNames {
        nameSet[n] = true
    }
    for _, t := range cfg.Tools {
        if t.Name != toolName {
            continue
        }
        adapter := getAdapter(t)
        candidates, _ := adapter.Pull(a.ctx, t.SkillsDir)
        for _, sk := range candidates {
            if !nameSet[sk.Name] {
                continue
            }
            existing, _ := a.storage.ListAll()
            for _, e := range existing {
                if e.Name == sk.Name {
                    _ = a.storage.Delete(e.ID)
                    break
                }
            }
            _, _ = a.storage.Import(sk.Path, category, skill.SourceManual, "", "")
        }
        go a.autoBackup()
    }
    return nil
}

func getAdapter(t config.ToolConfig) toolsync.ToolAdapter {
    if a, ok := registry.GetAdapter(t.Name); ok {
        return a
    }
    return toolsync.NewFilesystemAdapter(t.Name, t.SkillsDir)
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
    return a.config.Load()
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
    return a.config.Save(cfg)
}

func (a *App) AddCustomTool(name, skillsDir string) error {
    cfg, err := a.config.Load()
    if err != nil {
        return err
    }
    cfg.Tools = append(cfg.Tools, config.ToolConfig{
        Name:      name,
        SkillsDir: skillsDir,
        Enabled:   true,
        Custom:    true,
    })
    return a.config.Save(cfg)
}

func (a *App) RemoveCustomTool(name string) error {
    cfg, err := a.config.Load()
    if err != nil {
        return err
    }
    filtered := cfg.Tools[:0]
    for _, t := range cfg.Tools {
        if !(t.Custom && t.Name == name) {
            filtered = append(filtered, t)
        }
    }
    cfg.Tools = filtered
    return a.config.Save(cfg)
}

// --- Backup ---

func (a *App) BackupNow() error {
    a.autoBackup()
    return nil
}

func (a *App) ListCloudFiles() ([]backup.RemoteFile, error) {
    cfg, err := a.config.Load()
    if err != nil {
        return nil, err
    }
    provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
    if !ok {
        return nil, fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
    }
    if err := provider.Init(cfg.Cloud.Credentials); err != nil {
        return nil, err
    }
    return provider.List(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath)
}

func (a *App) RestoreFromCloud() error {
    cfg, err := a.config.Load()
    if err != nil {
        return err
    }
    provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
    if !ok {
        return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
    }
    if err := provider.Init(cfg.Cloud.Credentials); err != nil {
        return err
    }
    return provider.Restore(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath, cfg.SkillsStorageDir)
}

// ListCloudProviders returns all registered provider names and their required credential fields.
func (a *App) ListCloudProviders() []map[string]any {
    var result []map[string]any
    for _, p := range registry.AllCloudProviders() {
        result = append(result, map[string]any{
            "name":   p.Name(),
            "fields": p.RequiredCredentials(),
        })
    }
    return result
}

// --- Updates ---

func (a *App) CheckUpdates() error {
    skills, err := a.storage.ListAll()
    if err != nil {
        return err
    }
    for _, sk := range skills {
        result, err := a.checker.Check(a.ctx, sk)
        if err != nil {
            continue
        }
        if result.HasUpdate {
            sk.LatestSHA = result.LatestSHA
            _ = a.storage.UpdateMeta(sk)
            a.hub.Publish(notify.Event{
                Type: notify.EventUpdateAvailable,
                Payload: notify.UpdateAvailablePayload{
                    SkillID:    sk.ID,
                    SkillName:  sk.Name,
                    CurrentSHA: sk.SourceSHA,
                    LatestSHA:  result.LatestSHA,
                },
            })
        }
    }
    return nil
}

// UpdateSkill re-downloads a GitHub skill and updates local files and SHA.
func (a *App) UpdateSkill(skillID string) error {
    sk, err := a.storage.Get(skillID)
    if err != nil {
        return err
    }
    inst := install.NewGitHubInstaller("")
    tmpDir := filepath.Join(os.TempDir(), "skillflow-update", sk.Name)
    defer os.RemoveAll(tmpDir)

    c := install.SkillCandidate{Name: sk.Name, Path: sk.SourceSubPath}
    if err := inst.DownloadTo(a.ctx, install.InstallSource{Type: "github", URI: sk.SourceURL}, c, tmpDir); err != nil {
        return err
    }
    if err := a.storage.OverwriteFromDir(skillID, tmpDir); err != nil {
        return err
    }
    sk.SourceSHA = sk.LatestSHA
    sk.LatestSHA = ""
    _ = a.storage.UpdateMeta(sk)
    go a.autoBackup()
    return nil
}

func (a *App) checkUpdatesOnStartup() {
    _ = a.CheckUpdates()
}
```

**Step 2: Implement `app/wails/events.go`**

```go
package main

import (
    "context"
    "encoding/json"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func forwardEvents(ctx context.Context, hub *notify.Hub) {
    ch := hub.Subscribe()
    for {
        select {
        case evt, ok := <-ch:
            if !ok {
                return
            }
            data, _ := json.Marshal(evt.Payload)
            runtime.EventsEmit(ctx, string(evt.Type), string(data))
        case <-ctx.Done():
            return
        }
    }
}
```

**Step 3: Commit**

```bash
git add app/wails/
git commit -m "feat: add Wails app layer with all method bindings and event forwarding"
```

---

## Phase 8: Frontend

### Task 15: Frontend Setup

**Files:**
- Modify: `frontend/src/main.tsx`
- Create: `frontend/src/store/index.ts`
- Create: `frontend/src/App.tsx`

**Step 1: Install frontend dependencies**

```bash
cd frontend
npm install zustand @radix-ui/react-dialog react-router-dom lucide-react
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

**Step 2: Configure Tailwind in `frontend/src/index.css`**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

**Step 3: Create main layout with sidebar navigation in `frontend/src/App.tsx`**

```tsx
import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen bg-gray-950 text-gray-100">
        <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col p-4 gap-1">
          <h1 className="text-lg font-bold mb-6 px-2">SkillFlow</h1>
          <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
          <p className="text-xs text-gray-500 px-2 mt-3 mb-1">同步管理</p>
          <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
          <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
          <div className="flex-1" />
          <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
          <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
        </aside>
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/sync/push" element={<SyncPush />} />
            <Route path="/sync/pull" element={<SyncPull />} />
            <Route path="/backup" element={<Backup />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

function NavItem({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      end
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
          isActive ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'
        }`
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}
```

**Step 4: Commit**

```bash
git add frontend/
git commit -m "feat: add frontend layout with sidebar navigation"
```

---

### Task 16: Dashboard Page

**Files:**
- Create: `frontend/src/pages/Dashboard.tsx`
- Create: `frontend/src/components/SkillCard.tsx`
- Create: `frontend/src/components/CategoryPanel.tsx`
- Create: `frontend/src/components/ContextMenu.tsx`
- Create: `frontend/src/components/GitHubInstallDialog.tsx`
- Create: `frontend/src/components/ConflictDialog.tsx`

**Step 1: Implement ContextMenu component**

```tsx
// frontend/src/components/ContextMenu.tsx
import { useEffect, useRef } from 'react'

interface MenuItem { label: string; onClick: () => void; danger?: boolean }
interface Props { x: number; y: number; items: MenuItem[]; onClose: () => void }

export default function ContextMenu({ x, y, items, onClose }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose()
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [onClose])

  return (
    <div
      ref={ref}
      style={{ position: 'fixed', top: y, left: x, zIndex: 9999 }}
      className="bg-gray-800 border border-gray-700 rounded-lg shadow-xl py-1 min-w-36"
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => { item.onClick(); onClose() }}
          className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-700 ${item.danger ? 'text-red-400' : 'text-gray-200'}`}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}
```

**Step 2: Implement ConflictDialog component**

```tsx
// frontend/src/components/ConflictDialog.tsx
interface Props {
  conflicts: string[]
  onOverwrite: (name: string) => void
  onSkip: (name: string) => void
  onDone: () => void
}

export default function ConflictDialog({ conflicts, onOverwrite, onSkip, onDone }: Props) {
  if (conflicts.length === 0) { onDone(); return null }
  const current = conflicts[0]
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-2xl p-6 w-96 border border-gray-700">
        <h3 className="text-base font-semibold mb-2">冲突检测</h3>
        <p className="text-sm text-gray-400 mb-6">
          <span className="text-white font-medium">{current}</span> 已存在，如何处理？
        </p>
        <div className="flex gap-3 justify-end">
          <button
            onClick={() => onSkip(current)}
            className="px-4 py-2 text-sm rounded-lg bg-gray-700 hover:bg-gray-600"
          >跳过</button>
          <button
            onClick={() => onOverwrite(current)}
            className="px-4 py-2 text-sm rounded-lg bg-indigo-600 hover:bg-indigo-500"
          >覆盖</button>
        </div>
      </div>
    </div>
  )
}
```

**Step 3: Implement SkillCard with right-click menu**

```tsx
// frontend/src/components/SkillCard.tsx
import { useState } from 'react'
import { Github, FolderOpen, RefreshCw } from 'lucide-react'
import ContextMenu from './ContextMenu'

interface Skill { id: string; name: string; category: string; source: 'github' | 'manual'; hasUpdate: boolean }
interface Props {
  skill: Skill
  categories: string[]
  onDelete: () => void
  onUpdate?: () => void
  onMoveCategory: (category: string) => void
}

export default function SkillCard({ skill, categories, onDelete, onUpdate, onMoveCategory }: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null)

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault()
    setMenu({ x: e.clientX, y: e.clientY })
  }

  const menuItems = [
    ...(skill.hasUpdate ? [{ label: '更新', onClick: () => onUpdate?.() }] : []),
    ...categories.filter(c => c !== skill.category).map(c => ({
      label: `移动到 ${c || '未分类'}`,
      onClick: () => onMoveCategory(c),
    })),
    { label: '删除', onClick: onDelete, danger: true },
  ]

  return (
    <>
      <div
        draggable
        onDragStart={e => e.dataTransfer.setData('skillId', skill.id)}
        onContextMenu={handleContextMenu}
        className="relative bg-gray-800 border border-gray-700 rounded-xl p-4 cursor-grab hover:border-indigo-500 transition-colors group"
      >
        {skill.hasUpdate && (
          <span className="absolute top-2 right-2 w-2.5 h-2.5 rounded-full bg-red-500" />
        )}
        <div className="flex items-center gap-2 mb-2">
          {skill.source === 'github'
            ? <Github size={14} className="text-gray-400" />
            : <FolderOpen size={14} className="text-gray-400" />}
          <span className={`text-xs px-1.5 py-0.5 rounded ${skill.source === 'github' ? 'bg-blue-900/50 text-blue-300' : 'text-gray-400'}`}>
            {skill.source}
          </span>
        </div>
        <p className="font-medium text-sm truncate">{skill.name}</p>
        <div className="mt-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
          {skill.hasUpdate && (
            <button onClick={onUpdate} className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1">
              <RefreshCw size={12} /> 更新
            </button>
          )}
          <button onClick={onDelete} className="text-xs text-red-400 hover:text-red-300 ml-auto">删除</button>
        </div>
      </div>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menuItems} onClose={() => setMenu(null)} />
      )}
    </>
  )
}
```

**Step 4: Implement CategoryPanel with drag-drop and right-click rename/delete**

```tsx
// frontend/src/components/CategoryPanel.tsx
import { useState } from 'react'
import { Plus } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { CreateCategory, RenameCategory, DeleteCategory } from '../../wailsjs/go/main/App'

interface Props {
  categories: string[]
  selected: string | null
  onSelect: (cat: string | null) => void
  onDrop: (skillId: string, category: string) => void
  onRefresh: () => void
}

export default function CategoryPanel({ categories, selected, onSelect, onDrop, onRefresh }: Props) {
  const [menu, setMenu] = useState<{ x: number; y: number; cat: string } | null>(null)
  const [renaming, setRenaming] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createName, setCreateName] = useState('')

  const handleDrop = (e: React.DragEvent, cat: string) => {
    e.preventDefault()
    const id = e.dataTransfer.getData('skillId')
    if (id) onDrop(id, cat)
  }

  return (
    <div className="w-48 flex-shrink-0 border-r border-gray-800 p-3 flex flex-col gap-0.5">
      {/* All */}
      <div
        onClick={() => onSelect(null)}
        onDragOver={e => e.preventDefault()}
        onDrop={e => handleDrop(e, '')}
        className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${selected === null ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'}`}
      >全部</div>

      {/* Categories */}
      {categories.map(cat => (
        renaming === cat
          ? <input
              key={cat} autoFocus value={newName}
              onChange={e => setNewName(e.target.value)}
              onBlur={async () => {
                if (newName && newName !== cat) { await RenameCategory(cat, newName); onRefresh() }
                setRenaming(null)
              }}
              onKeyDown={async e => {
                if (e.key === 'Enter') { await RenameCategory(cat, newName); onRefresh(); setRenaming(null) }
                if (e.key === 'Escape') setRenaming(null)
              }}
              className="px-3 py-1.5 rounded-lg text-sm bg-gray-700 text-white outline-none w-full"
            />
          : <div
              key={cat}
              onClick={() => onSelect(cat)}
              onDragOver={e => e.preventDefault()}
              onDrop={e => handleDrop(e, cat)}
              onContextMenu={e => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY, cat }) }}
              className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${selected === cat ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'}`}
            >{cat}</div>
      ))}

      {/* New category input */}
      {creating
        ? <input
            autoFocus value={createName}
            onChange={e => setCreateName(e.target.value)}
            onBlur={async () => {
              if (createName) { await CreateCategory(createName); onRefresh() }
              setCreating(false); setCreateName('')
            }}
            onKeyDown={async e => {
              if (e.key === 'Enter') { await CreateCategory(createName); onRefresh(); setCreating(false); setCreateName('') }
              if (e.key === 'Escape') { setCreating(false); setCreateName('') }
            }}
            className="px-3 py-1.5 rounded-lg text-sm bg-gray-700 text-white outline-none w-full"
          />
        : <button
            onClick={() => setCreating(true)}
            className="flex items-center gap-1.5 px-3 py-2 text-sm text-gray-500 hover:text-gray-300 mt-1"
          ><Plus size={14} /> 新建分类</button>
      }

      {/* Context menu */}
      {menu && (
        <ContextMenu
          x={menu.x} y={menu.y}
          items={[
            { label: '重命名', onClick: () => { setRenaming(menu.cat); setNewName(menu.cat) } },
            { label: '删除', onClick: async () => { await DeleteCategory(menu.cat); onRefresh() }, danger: true },
          ]}
          onClose={() => setMenu(null)}
        />
      )}
    </div>
  )
}
```

**Step 5: Implement GitHubInstallDialog**

```tsx
// frontend/src/components/GitHubInstallDialog.tsx
import { useState } from 'react'
import { ScanGitHub, InstallFromGitHub, ListCategories } from '../../wailsjs/go/main/App'
import { Github, X } from 'lucide-react'

interface Props { onClose: () => void; onDone: () => void }

export default function GitHubInstallDialog({ onClose, onDone }: Props) {
  const [url, setUrl] = useState('')
  const [candidates, setCandidates] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [category, setCategory] = useState('')
  const [scanning, setScanning] = useState(false)
  const [installing, setInstalling] = useState(false)

  const scan = async () => {
    setScanning(true)
    const [c, cats] = await Promise.all([ScanGitHub(url), ListCategories()])
    setCandidates(c ?? [])
    setCategories(cats ?? [])
    setSelected(new Set((c ?? []).filter((x: any) => !x.Installed).map((x: any) => x.Name)))
    setScanning(false)
  }

  const install = async () => {
    setInstalling(true)
    const toInstall = candidates.filter(c => selected.has(c.Name))
    await InstallFromGitHub(url, toInstall, category)
    setInstalling(false)
    onDone()
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-2xl p-6 w-[520px] border border-gray-700">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold flex items-center gap-2"><Github size={16} /> 从 GitHub 安装</h3>
          <button onClick={onClose}><X size={16} className="text-gray-400" /></button>
        </div>

        <div className="flex gap-2 mb-4">
          <input
            value={url} onChange={e => setUrl(e.target.value)}
            placeholder="https://github.com/user/repo"
            className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500"
          />
          <button onClick={scan} disabled={scanning || !url} className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
            {scanning ? '扫描中...' : '扫描'}
          </button>
        </div>

        {candidates.length > 0 && (
          <>
            <div className="max-h-52 overflow-y-auto space-y-1 mb-4">
              {candidates.map(c => (
                <label key={c.Name} className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-700 cursor-pointer">
                  <input type="checkbox" checked={selected.has(c.Name)} onChange={() => toggle(c.Name)} className="accent-indigo-500" />
                  <span className="text-sm flex-1">{c.Name}</span>
                  {c.Installed && <span className="text-xs bg-blue-900/50 text-blue-300 px-2 py-0.5 rounded">已安装</span>}
                </label>
              ))}
            </div>
            <div className="flex items-center gap-3 mb-4">
              <span className="text-sm text-gray-400">安装到分类</span>
              <select
                value={category} onChange={e => setCategory(e.target.value)}
                className="bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm flex-1"
              >
                <option value="">未分类</option>
                {categories.map(c => <option key={c} value={c}>{c}</option>)}
              </select>
            </div>
            <button
              onClick={install} disabled={installing || selected.size === 0}
              className="w-full py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
            >{installing ? '安装中...' : `安装 ${selected.size} 个 Skill`}</button>
          </>
        )}
      </div>
    </div>
  )
}
```

**Step 6: Implement Dashboard page with toolbar and window drag-drop**

```tsx
// frontend/src/pages/Dashboard.tsx
import { useEffect, useState, useCallback } from 'react'
import {
  ListSkills, ListCategories, MoveSkillCategory,
  DeleteSkill, ImportLocal, UpdateSkill, CheckUpdates
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'
import { runtime } from '../../wailsjs/runtime'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'
import GitHubInstallDialog from '../components/GitHubInstallDialog'
import { Github, FolderOpen, RefreshCw, Search } from 'lucide-react'

export default function Dashboard() {
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showGitHub, setShowGitHub] = useState(false)
  const [dragOver, setDragOver] = useState(false)

  const load = useCallback(async () => {
    const [s, c] = await Promise.all([ListSkills(), ListCategories()])
    setSkills(s ?? [])
    setCategories(c ?? [])
  }, [])

  useEffect(() => {
    load()
    // Listen for update-available events from backend
    EventsOn('update.available', load)
  }, [load])

  const filtered = skills.filter(sk => {
    const matchCat = selectedCat === null || sk.Category === selectedCat
    const matchSearch = !search || sk.Name.toLowerCase().includes(search.toLowerCase())
    return matchCat && matchSearch
  })

  const handleDrop = async (skillId: string, category: string) => {
    await MoveSkillCategory(skillId, category)
    load()
  }

  // Window-level drag-drop: import a folder dropped anywhere on the page
  const handleWindowDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }
  const handleWindowDragLeave = () => setDragOver(false)
  const handleWindowDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const items = Array.from(e.dataTransfer.items)
    for (const item of items) {
      const entry = item.webkitGetAsEntry?.()
      if (entry?.isDirectory) {
        // Wails exposes the local path via the file object
        const file = item.getAsFile()
        if (file) {
          // @ts-ignore — Wails provides .path on File objects
          await ImportLocal(file.path ?? file.name, selectedCat ?? '')
          load()
        }
      }
    }
  }

  const handleImportButton = async () => {
    const dir = await runtime.OpenDirectoryDialog({ Title: '选择 Skill 目录' })
    if (dir) { await ImportLocal(dir, selectedCat ?? ''); load() }
  }

  return (
    <div
      className={`flex h-full relative ${dragOver ? 'ring-2 ring-inset ring-indigo-500' : ''}`}
      onDragOver={handleWindowDragOver}
      onDragLeave={handleWindowDragLeave}
      onDrop={handleWindowDrop}
    >
      {dragOver && (
        <div className="absolute inset-0 bg-indigo-500/10 flex items-center justify-center z-40 pointer-events-none">
          <p className="text-indigo-300 text-lg font-medium">松开以导入 Skill</p>
        </div>
      )}

      <CategoryPanel
        categories={categories}
        selected={selectedCat}
        onSelect={setSelectedCat}
        onDrop={handleDrop}
        onRefresh={load}
      />

      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center gap-3 px-6 py-4 border-b border-gray-800">
          <div className="relative flex-1 max-w-xs">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input
              value={search} onChange={e => setSearch(e.target.value)}
              placeholder="搜索 Skills..."
              className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-8 pr-3 py-1.5 text-sm outline-none focus:border-indigo-500"
            />
          </div>
          <button
            onClick={() => CheckUpdates()}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
          ><RefreshCw size={14} /> 检查更新</button>
          <button
            onClick={handleImportButton}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800"
          ><FolderOpen size={14} /> 手动导入</button>
          <button
            onClick={() => setShowGitHub(true)}
            className="flex items-center gap-1.5 px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 rounded-lg"
          ><Github size={14} /> 从 GitHub 安装</button>
        </div>

        {/* Skills grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
            {filtered.map(sk => (
              <SkillCard
                key={sk.ID}
                skill={{ id: sk.ID, name: sk.Name, category: sk.Category, source: sk.Source, hasUpdate: !!sk.LatestSHA }}
                categories={categories}
                onDelete={async () => { await DeleteSkill(sk.ID); load() }}
                onUpdate={async () => { await UpdateSkill(sk.ID); load() }}
                onMoveCategory={async cat => { await MoveSkillCategory(sk.ID, cat); load() }}
              />
            ))}
          </div>
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <p className="text-sm">没有找到 Skills</p>
              <p className="text-xs mt-1">从 GitHub 安装或拖拽文件夹到此处</p>
            </div>
          )}
        </div>
      </div>

      {showGitHub && (
        <GitHubInstallDialog onClose={() => setShowGitHub(false)} onDone={() => { setShowGitHub(false); load() }} />
      )}
    </div>
  )
}
```

**Step 7: Commit**

```bash
git add frontend/src/
git commit -m "feat: add Dashboard with full toolbar, drag-drop, right-click menus, GitHub install dialog"
```

---

### Task 17: Sync Push and Pull Pages

**Files:**
- Create: `frontend/src/pages/SyncPush.tsx`
- Create: `frontend/src/pages/SyncPull.tsx`

**Step 1: Implement SyncPush.tsx**

```tsx
// frontend/src/pages/SyncPush.tsx
import { useEffect, useState } from 'react'
import { GetEnabledTools, ListSkills, ListCategories, PushToTools, PushToToolsForce } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import { ArrowUpFromLine } from 'lucide-react'

type Scope = 'all' | 'category' | 'manual'

export default function SyncPush() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [scope, setScope] = useState<Scope>('all')
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState('')
  const [skills, setSkills] = useState<any[]>([])
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [conflicts, setConflicts] = useState<string[]>([])
  const [pushing, setPushing] = useState(false)
  const [done, setDone] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListSkills(), ListCategories()]).then(([t, s, c]) => {
      setTools(t ?? [])
      setSkills(s ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const getSkillIDs = () => {
    if (scope === 'all') return skills.map(s => s.ID)
    if (scope === 'category') return skills.filter(s => s.Category === selectedCategory).map(s => s.ID)
    return [...selectedSkills]
  }

  const push = async () => {
    setPushing(true)
    setDone(false)
    const ids = getSkillIDs()
    const toolNames = [...selectedTools]
    const conflicts = await PushToTools(ids, toolNames)
    if (conflicts && conflicts.length > 0) {
      setConflicts(conflicts)
    } else {
      setDone(true)
    }
    setPushing(false)
  }

  const toggleTool = (name: string) => {
    const next = new Set(selectedTools)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelectedTools(next)
  }

  const toggleSkill = (id: string) => {
    const next = new Set(selectedSkills)
    next.has(id) ? next.delete(id) : next.add(id)
    setSelectedSkills(next)
  }

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><ArrowUpFromLine size={18} /> 推送到工具</h2>

      {/* Tool selection */}
      <section className="mb-6">
        <p className="text-sm text-gray-400 mb-3">目标工具</p>
        <div className="flex flex-wrap gap-2">
          {tools.map(t => (
            <button
              key={t.Name}
              onClick={() => toggleTool(t.Name)}
              className={`px-4 py-2 rounded-lg text-sm border transition-colors ${selectedTools.has(t.Name) ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'}`}
            >{t.Name}</button>
          ))}
        </div>
      </section>

      {/* Scope selection */}
      <section className="mb-6">
        <p className="text-sm text-gray-400 mb-3">同步范围</p>
        <div className="space-y-2">
          {([['all', '全部 Skills'], ['category', '按分类'], ['manual', '手动选择']] as [Scope, string][]).map(([v, label]) => (
            <label key={v} className="flex items-center gap-3 cursor-pointer">
              <input type="radio" checked={scope === v} onChange={() => setScope(v)} className="accent-indigo-500" />
              <span className="text-sm">{label}</span>
            </label>
          ))}
        </div>

        {scope === 'category' && (
          <select value={selectedCategory} onChange={e => setSelectedCategory(e.target.value)}
            className="mt-3 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm w-48">
            <option value="">选择分类</option>
            {categories.map(c => <option key={c} value={c}>{c}</option>)}
          </select>
        )}

        {scope === 'manual' && (
          <div className="mt-3 max-h-52 overflow-y-auto space-y-1 border border-gray-700 rounded-xl p-3">
            {skills.map(sk => (
              <label key={sk.ID} className="flex items-center gap-3 px-2 py-1.5 hover:bg-gray-800 rounded-lg cursor-pointer">
                <input type="checkbox" checked={selectedSkills.has(sk.ID)} onChange={() => toggleSkill(sk.ID)} className="accent-indigo-500" />
                <span className="text-sm">{sk.Name}</span>
                <span className="text-xs text-gray-500">{sk.Category || '未分类'}</span>
              </label>
            ))}
          </div>
        )}
      </section>

      <button
        onClick={push}
        disabled={pushing || selectedTools.size === 0}
        className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
      >{pushing ? '推送中...' : '开始推送'}</button>

      {done && <p className="mt-4 text-sm text-green-400">推送完成</p>}

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            const skill = skills.find(s => s.Name === name)
            if (skill) await PushToToolsForce([skill.ID], [...selectedTools])
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
```

**Step 2: Implement SyncPull.tsx**

```tsx
// frontend/src/pages/SyncPull.tsx
import { useEffect, useState } from 'react'
import { GetEnabledTools, ScanToolSkills, PullFromTool, PullFromToolForce, ListCategories } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import { ArrowDownToLine } from 'lucide-react'

export default function SyncPull() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTool, setSelectedTool] = useState('')
  const [scanned, setScanned] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [targetCategory, setTargetCategory] = useState('')
  const [scanning, setScanning] = useState(false)
  const [pulling, setPulling] = useState(false)
  const [conflicts, setConflicts] = useState<string[]>([])
  const [done, setDone] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListCategories()]).then(([t, c]) => {
      setTools(t ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const scan = async () => {
    setScanning(true)
    setScanned([])
    const skills = await ScanToolSkills(selectedTool)
    setScanned(skills ?? [])
    setSelected(new Set((skills ?? []).map((s: any) => s.Name)))
    setScanning(false)
  }

  const pull = async () => {
    setPulling(true)
    const names = [...selected]
    const conflicts = await PullFromTool(selectedTool, names, targetCategory)
    if (conflicts && conflicts.length > 0) {
      setConflicts(conflicts)
    } else {
      setDone(true)
    }
    setPulling(false)
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><ArrowDownToLine size={18} /> 从工具拉取</h2>

      {/* Tool select */}
      <section className="mb-4">
        <p className="text-sm text-gray-400 mb-3">来源工具</p>
        <div className="flex flex-wrap gap-2">
          {tools.map(t => (
            <button
              key={t.Name}
              onClick={() => { setSelectedTool(t.Name); setScanned([]); setDone(false) }}
              className={`px-4 py-2 rounded-lg text-sm border transition-colors ${selectedTool === t.Name ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'}`}
            >{t.Name}</button>
          ))}
        </div>
      </section>

      <button
        onClick={scan} disabled={!selectedTool || scanning}
        className="mb-6 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm disabled:opacity-50"
      >{scanning ? '扫描中...' : '扫描'}</button>

      {scanned.length > 0 && (
        <>
          <section className="mb-4">
            <p className="text-sm text-gray-400 mb-2">选择要导入的 Skills（{selected.size}/{scanned.length}）</p>
            <div className="max-h-52 overflow-y-auto space-y-1 border border-gray-700 rounded-xl p-3">
              {scanned.map(sk => (
                <label key={sk.Name} className="flex items-center gap-3 px-2 py-1.5 hover:bg-gray-800 rounded-lg cursor-pointer">
                  <input type="checkbox" checked={selected.has(sk.Name)} onChange={() => toggle(sk.Name)} className="accent-indigo-500" />
                  <span className="text-sm">{sk.Name}</span>
                </label>
              ))}
            </div>
          </section>

          <section className="mb-6 flex items-center gap-3">
            <span className="text-sm text-gray-400">导入到分类</span>
            <select value={targetCategory} onChange={e => setTargetCategory(e.target.value)}
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-sm">
              <option value="">Imported（默认）</option>
              {categories.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </section>

          <button
            onClick={pull} disabled={pulling || selected.size === 0}
            className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
          >{pulling ? '拉取中...' : '开始拉取'}</button>

          {done && <p className="mt-4 text-sm text-green-400">拉取完成</p>}
        </>
      )}

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            await PullFromToolForce(selectedTool, [name], targetCategory)
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add SyncPush and SyncPull pages with conflict dialog"
```

---

### Task 18: Backup and Settings Pages

**Files:**
- Create: `frontend/src/pages/Backup.tsx`
- Create: `frontend/src/pages/Settings.tsx`

**Step 1: Implement Backup.tsx**

```tsx
// frontend/src/pages/Backup.tsx
import { useEffect, useState } from 'react'
import { BackupNow, ListCloudFiles, RestoreFromCloud, GetConfig } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'
import { Cloud, Upload, Download, RefreshCw } from 'lucide-react'

export default function Backup() {
  const [files, setFiles] = useState<any[]>([])
  const [status, setStatus] = useState<'idle' | 'backing-up' | 'done' | 'error'>('idle')
  const [currentFile, setCurrentFile] = useState('')
  const [cloudEnabled, setCloudEnabled] = useState(false)

  useEffect(() => {
    GetConfig().then(cfg => setCloudEnabled(cfg?.Cloud?.Enabled ?? false))
    EventsOn('backup.started', () => setStatus('backing-up'))
    EventsOn('backup.progress', (data: string) => {
      try { setCurrentFile(JSON.parse(data).currentFile ?? '') } catch {}
    })
    EventsOn('backup.completed', () => { setStatus('done'); loadFiles() })
    EventsOn('backup.failed', () => setStatus('error'))
  }, [])

  const loadFiles = async () => {
    const f = await ListCloudFiles()
    setFiles(f ?? [])
  }

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><Cloud size={18} /> 云备份</h2>

      {!cloudEnabled && (
        <div className="bg-yellow-900/30 border border-yellow-700/50 rounded-xl p-4 mb-6 text-sm text-yellow-300">
          云备份未启用。请前往设置 → 云存储完成配置。
        </div>
      )}

      <div className="flex gap-3 mb-8">
        <button
          onClick={async () => { await BackupNow() }}
          disabled={!cloudEnabled || status === 'backing-up'}
          className="flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
        >
          {status === 'backing-up' ? <RefreshCw size={14} className="animate-spin" /> : <Upload size={14} />}
          {status === 'backing-up' ? `备份中 ${currentFile}` : '立即备份'}
        </button>
        <button
          onClick={async () => { await RestoreFromCloud(); loadFiles() }}
          disabled={!cloudEnabled}
          className="flex items-center gap-2 px-5 py-2.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm disabled:opacity-50"
        ><Download size={14} /> 从云端恢复</button>
        <button onClick={loadFiles} className="flex items-center gap-2 px-4 py-2.5 text-gray-400 hover:text-white rounded-lg hover:bg-gray-800 text-sm">
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      {status === 'done' && <p className="mb-4 text-sm text-green-400">备份完成</p>}
      {status === 'error' && <p className="mb-4 text-sm text-red-400">备份失败，请检查云存储配置</p>}

      {files.length > 0 && (
        <div>
          <p className="text-sm text-gray-400 mb-3">云端文件（{files.length} 个）</p>
          <div className="max-h-96 overflow-y-auto border border-gray-800 rounded-xl divide-y divide-gray-800">
            {files.map((f, i) => (
              <div key={i} className="flex items-center justify-between px-4 py-2.5 text-sm">
                <span className="text-gray-300 font-mono text-xs">{f.Path}</span>
                <span className="text-gray-500 text-xs">{(f.Size / 1024).toFixed(1)} KB</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
```

**Step 2: Implement Settings.tsx**

```tsx
// frontend/src/pages/Settings.tsx
import { useEffect, useState } from 'react'
import { GetConfig, SaveConfig, ListCloudProviders, AddCustomTool, RemoveCustomTool } from '../../wailsjs/go/main/App'
import { Plus, Trash2, Settings } from 'lucide-react'

type Tab = 'tools' | 'cloud' | 'general'

export default function SettingsPage() {
  const [tab, setTab] = useState<Tab>('tools')
  const [cfg, setCfg] = useState<any>(null)
  const [providers, setProviders] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [newTool, setNewTool] = useState({ name: '', skillsDir: '' })

  useEffect(() => {
    Promise.all([GetConfig(), ListCloudProviders()]).then(([c, p]) => {
      setCfg(c)
      setProviders(p ?? [])
    })
  }, [])

  const save = async () => {
    setSaving(true)
    await SaveConfig(cfg)
    setSaving(false)
  }

  const updateTool = (name: string, field: string, value: any) => {
    setCfg((prev: any) => ({
      ...prev,
      tools: prev.tools.map((t: any) => t.name === name ? { ...t, [field]: value } : t)
    }))
  }

  const selectedProvider = providers.find((p: any) => p.name === cfg?.cloud?.provider)

  if (!cfg) return <div className="p-8 text-gray-400">加载中...</div>

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><Settings size={18} /> 设置</h2>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-gray-800 rounded-xl p-1 w-fit">
        {([['tools', '工具路径'], ['cloud', '云存储'], ['general', '通用']] as [Tab, string][]).map(([v, label]) => (
          <button key={v} onClick={() => setTab(v)}
            className={`px-4 py-1.5 rounded-lg text-sm transition-colors ${tab === v ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'}`}
          >{label}</button>
        ))}
      </div>

      {/* Tools tab */}
      {tab === 'tools' && (
        <div className="space-y-4">
          {cfg.tools.map((t: any) => (
            <div key={t.name} className="bg-gray-800 rounded-xl p-4 border border-gray-700">
              <div className="flex items-center justify-between mb-3">
                <span className="font-medium text-sm">{t.name}</span>
                <label className="flex items-center gap-2 cursor-pointer">
                  <span className="text-xs text-gray-400">启用</span>
                  <div
                    onClick={() => updateTool(t.name, 'enabled', !t.enabled)}
                    className={`w-9 h-5 rounded-full transition-colors relative ${t.enabled ? 'bg-indigo-600' : 'bg-gray-600'}`}
                  >
                    <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${t.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
                  </div>
                </label>
              </div>
              <input
                value={t.skillsDir}
                onChange={e => updateTool(t.name, 'skillsDir', e.target.value)}
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none focus:border-indigo-500"
              />
              {t.custom && (
                <button
                  onClick={async () => { await RemoveCustomTool(t.name); const c = await GetConfig(); setCfg(c) }}
                  className="mt-2 text-xs text-red-400 hover:text-red-300 flex items-center gap-1"
                ><Trash2 size={12} /> 删除</button>
              )}
            </div>
          ))}

          {/* Add custom tool */}
          <div className="bg-gray-800 rounded-xl p-4 border border-dashed border-gray-600">
            <p className="text-sm text-gray-400 mb-3">添加自定义工具</p>
            <div className="flex gap-2 mb-2">
              <input value={newTool.name} onChange={e => setNewTool(p => ({ ...p, name: e.target.value }))}
                placeholder="工具名称" className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm outline-none" />
            </div>
            <div className="flex gap-2">
              <input value={newTool.skillsDir} onChange={e => setNewTool(p => ({ ...p, skillsDir: e.target.value }))}
                placeholder="/path/to/skills" className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm font-mono outline-none" />
              <button
                onClick={async () => {
                  if (newTool.name && newTool.skillsDir) {
                    await AddCustomTool(newTool.name, newTool.skillsDir)
                    const c = await GetConfig(); setCfg(c)
                    setNewTool({ name: '', skillsDir: '' })
                  }
                }}
                className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm flex items-center gap-1"
              ><Plus size={14} /> 添加</button>
            </div>
          </div>
        </div>
      )}

      {/* Cloud tab */}
      {tab === 'cloud' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-400 mb-2">云厂商</p>
            <div className="flex gap-2">
              {providers.map((p: any) => (
                <button key={p.name}
                  onClick={() => setCfg((prev: any) => ({ ...prev, cloud: { ...prev.cloud, provider: p.name } }))}
                  className={`px-4 py-2 rounded-lg text-sm border transition-colors ${cfg.cloud?.provider === p.name ? 'bg-indigo-600 border-indigo-500' : 'bg-gray-800 border-gray-700 hover:border-gray-500'}`}
                >{p.name}</button>
              ))}
            </div>
          </div>

          {selectedProvider && (
            <>
              <div>
                <p className="text-sm text-gray-400 mb-2">存储桶</p>
                <input value={cfg.cloud?.bucketName ?? ''} onChange={e => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, bucketName: e.target.value } }))}
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
              </div>
              {/* Dynamic credential fields from RequiredCredentials() */}
              {selectedProvider.fields.map((f: any) => (
                <div key={f.key}>
                  <p className="text-sm text-gray-400 mb-2">{f.label}</p>
                  <input
                    type={f.secret ? 'password' : 'text'}
                    placeholder={f.placeholder ?? ''}
                    value={cfg.cloud?.credentials?.[f.key] ?? ''}
                    onChange={e => setCfg((p: any) => ({
                      ...p, cloud: { ...p.cloud, credentials: { ...p.cloud?.credentials, [f.key]: e.target.value } }
                    }))}
                    className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500 font-mono"
                  />
                </div>
              ))}
              <label className="flex items-center gap-3 cursor-pointer">
                <div
                  onClick={() => setCfg((p: any) => ({ ...p, cloud: { ...p.cloud, enabled: !p.cloud?.enabled } }))}
                  className={`w-9 h-5 rounded-full transition-colors relative ${cfg.cloud?.enabled ? 'bg-indigo-600' : 'bg-gray-600'}`}
                >
                  <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${cfg.cloud?.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
                </div>
                <span className="text-sm text-gray-300">启用自动云备份</span>
              </label>
            </>
          )}
        </div>
      )}

      {/* General tab */}
      {tab === 'general' && (
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-400 mb-2">本地 Skills 存储目录</p>
            <input value={cfg.skillsStorageDir ?? ''} onChange={e => setCfg((p: any) => ({ ...p, skillsStorageDir: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm font-mono outline-none focus:border-indigo-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400 mb-2">从工具拉取时的默认分类</p>
            <input value={cfg.defaultCategory ?? ''} onChange={e => setCfg((p: any) => ({ ...p, defaultCategory: e.target.value }))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
          </div>
        </div>
      )}

      <button onClick={save} disabled={saving}
        className="mt-8 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
        {saving ? '保存中...' : '保存设置'}
      </button>
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add Backup and Settings pages with dynamic cloud form and custom tool management"
```

---

## Phase 9: Build & CI

### Task 19: Build Configuration and GitHub Actions

**Files:**
- Create: `.github/workflows/build.yml`
- Modify: `wails.json`

**Step 1: Configure wails.json**

```json
{
  "schemaVersion": "2",
  "name": "SkillFlow",
  "outputfilename": "SkillFlow",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto"
}
```

**Step 2: Create `.github/workflows/build.yml`**

```yaml
name: Build
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: macos-latest
            arch: amd64
            name: macos-intel
          - os: macos-latest
            arch: arm64
            name: macos-apple-silicon
          - os: windows-latest
            arch: amd64
            name: windows

    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: wails build -platform ${{ matrix.arch == 'arm64' && 'darwin/arm64' || matrix.os == 'windows-latest' && 'windows/amd64' || 'darwin/amd64' }}
      - uses: actions/upload-artifact@v4
        with:
          name: skillflow-${{ matrix.name }}
          path: build/bin/
```

**Step 3: Final local build test**

```bash
wails build
```
Expected: Binary produced in `build/bin/`

**Step 4: Commit**

```bash
git add .github/ wails.json
git commit -m "chore: add GitHub Actions matrix build for macOS (intel+arm) and Windows"
```

---

## Summary

| Phase | Tasks | Key Deliverables |
|-------|-------|-----------------|
| 1 Foundation | 1-5 | Go module, Wails init, models, notify hub, config, registry |
| 2 Skill Mgmt | 6-7 | Validator, storage with categories and meta |
| 3 Install | 8-10 | Install interface, GitHub installer, local installer |
| 4 Sync | 11 | Filesystem adapter (shared by all tools), registry wiring |
| 5 Backup | 12 | Cloud provider interface + Aliyun/Tencent/Huawei |
| 6 Update | 13 | GitHub SHA-based update checker |
| 7 App Layer | 14 | All Wails method bindings + event forwarding |
| 8 Frontend | 15-18 | All 5 pages + drag-drop + dynamic forms |
| 9 Build | 19 | GitHub Actions matrix build |
