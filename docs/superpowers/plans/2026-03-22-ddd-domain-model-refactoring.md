# DDD Domain Model Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor all bounded context domain packages to introduce the value objects, aggregate roots, and published language types described in `docs/architecture/contexts.md`, closing the gap between the architecture document and the actual code.

**Architecture:** Each bounded context gets its own value objects file (`value_objects.go`) and, where needed, new aggregate root files. Value objects that are simple identity types use `type X string`. Composite value objects use small structs. Published language types are defined in each context's domain package. The shared kernel (`core/shared/`) gains a `LogicalSkillKey` named type, domain error contracts, and base event contracts.

**Tech Stack:** Go 1.25, Wails v2 backend bindings, testify for Go tests

**Patch to Preserve:** Commit `22d5c2e` (refactor: keep synced skill data under app data root) must not be broken. All changes must be verified against existing tests via `go test ./core/... ./cmd/skillflow`.

**Two-phase approach:** This plan (Phase 1) introduces all types and wires them via accessor methods on existing aggregates, keeping existing struct field types as `string` to avoid cascading compile breakage. A follow-up plan (Phase 2) will migrate aggregate field types from `string` to named types across the entire codebase once the types are proven stable.

**Task parallelism:** Task 1 is a prerequisite for all others. After Task 1 completes, Tasks 2–6 are independent and can run in parallel. Task 7 runs after all others.

---

## Task 1: Shared Kernel — `LogicalSkillKey` Type, Domain Errors, Base Events

**Files:**
- Modify: `core/shared/logicalkey/logicalkey.go`
- Modify: `core/shared/logicalkey/logicalkey_test.go` (add type-related tests)
- Create: `core/shared/domainerr/errors.go`
- Create: `core/shared/domainerr/errors_test.go`
- Create: `core/shared/domainevent/event.go`

### 1.1 `LogicalSkillKey` named type

- [ ] **Step 1: Write failing test for LogicalSkillKey type**

```go
// core/shared/logicalkey/logicalkey_test.go (add to existing file)
func TestLogicalSkillKeyStringConversion(t *testing.T) {
	key := logicalkey.LogicalSkillKey("git:github.com/user/repo#path")
	assert.Equal(t, "git:github.com/user/repo#path", key.String())
	assert.False(t, key.IsEmpty())

	empty := logicalkey.LogicalSkillKey("")
	assert.True(t, empty.IsEmpty())
}

func TestGitReturnsLogicalSkillKey(t *testing.T) {
	key := logicalkey.Git("github.com/user/repo", "skills/my-skill")
	assert.IsType(t, logicalkey.LogicalSkillKey(""), key)
	assert.Equal(t, "git:github.com/user/repo#skills/my-skill", key.String())
}

func TestContentFromDirReturnsLogicalSkillKey(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# test"), 0644))
	key, err := logicalkey.ContentFromDir(dir)
	require.NoError(t, err)
	assert.IsType(t, logicalkey.LogicalSkillKey(""), key)
	assert.False(t, key.IsEmpty())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/shared/logicalkey/ -run TestLogicalSkillKeyStringConversion -v`
Expected: FAIL — `LogicalSkillKey` type and methods not defined

- [ ] **Step 3: Implement LogicalSkillKey type and update function signatures**

In `core/shared/logicalkey/logicalkey.go`:

```go
// LogicalSkillKey is the cross-context identity for a logical skill.
// For repo-backed skills: "git:<repo-source>#<subpath>"
// For manual skills: "content:<sha256-hash>"
type LogicalSkillKey string

func (k LogicalSkillKey) String() string { return string(k) }
func (k LogicalSkillKey) IsEmpty() bool  { return strings.TrimSpace(string(k)) == "" }
```

Update return types of `Git`, `GitFromRepoURL`, `GitFromRepoURLOrEmpty`, `ContentFromDir` from `string` to `LogicalSkillKey`.

Then fix all callers across the codebase. Since `LogicalSkillKey` is `type X string`, most sites just need `string(key)` conversions where a plain `string` is expected. Key caller sites:
- `core/skillcatalog/domain/installed_skill.go` — `LogicalKey()` returns `(LogicalSkillKey, error)`
- `core/skillcatalog/app/query/installed_index.go` — `installedGroup.LogicalKey` field type → `string` stays, use `string(key)` at assignment
- `core/agentintegration/app/service.go` — `SkillStatus.LogicalKey` field stays `string`, convert at call sites
- `core/readmodel/skills/composer.go` — convert at assignment
- `cmd/skillflow/` transport layer — convert at assignment

Strategy: keep all struct fields that carry logical keys as `string` in this plan phase. Only the function return types and local variables change. This avoids cascading JSON or struct literal breakage.

- [ ] **Step 4: Run all tests**

Run: `go test ./core/... ./cmd/skillflow`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add core/shared/logicalkey/
git commit -m "refactor(shared): introduce LogicalSkillKey named type"
```

### 1.2 Domain error contracts

- [ ] **Step 6: Write failing test**

```go
// core/shared/domainerr/errors_test.go
package domainerr_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/shared/domainerr"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	err := domainerr.NotFound("skill", "abc-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skill")
	assert.Contains(t, err.Error(), "abc-123")
	assert.True(t, domainerr.IsNotFound(err))
}

func TestAlreadyExistsError(t *testing.T) {
	err := domainerr.AlreadyExists("skill", "my-skill")
	assert.Error(t, err)
	assert.True(t, domainerr.IsAlreadyExists(err))
}

func TestValidationError(t *testing.T) {
	err := domainerr.Validation("name must not be empty")
	assert.Error(t, err)
	assert.True(t, domainerr.IsValidation(err))
}
```

- [ ] **Step 7: Run test to verify it fails**

Run: `go test ./core/shared/domainerr/ -v`
Expected: FAIL — package does not exist

- [ ] **Step 8: Implement domain error contracts**

```go
// core/shared/domainerr/errors.go
package domainerr

import (
	"errors"
	"fmt"
)

type notFoundError struct{ msg string }
func (e *notFoundError) Error() string { return e.msg }

type alreadyExistsError struct{ msg string }
func (e *alreadyExistsError) Error() string { return e.msg }

type validationError struct{ msg string }
func (e *validationError) Error() string { return e.msg }

func NotFound(entity, id string) error {
	return &notFoundError{msg: fmt.Sprintf("%s not found: %s", entity, id)}
}

func AlreadyExists(entity, id string) error {
	return &alreadyExistsError{msg: fmt.Sprintf("%s already exists: %s", entity, id)}
}

func Validation(msg string) error {
	return &validationError{msg: msg}
}

func IsNotFound(err error) bool {
	var target *notFoundError
	return errors.As(err, &target)
}

func IsAlreadyExists(err error) bool {
	var target *alreadyExistsError
	return errors.As(err, &target)
}

func IsValidation(err error) bool {
	var target *validationError
	return errors.As(err, &target)
}
```

- [ ] **Step 9: Run test to verify it passes**

Run: `go test ./core/shared/domainerr/ -v`
Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add core/shared/domainerr/
git commit -m "refactor(shared): add domain error contracts"
```

### 1.3 Base domain event contracts

- [ ] **Step 11: Create base event contract**

```go
// core/shared/domainevent/event.go
package domainevent

import "time"

// Event is the base contract for domain events published by bounded contexts.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Base provides common fields for domain events.
type Base struct {
	Name       string    `json:"eventName"`
	OccurredTs time.Time `json:"occurredAt"`
}

func (b Base) EventName() string    { return b.Name }
func (b Base) OccurredAt() time.Time { return b.OccurredTs }

func NewBase(name string) Base {
	return Base{Name: name, OccurredTs: time.Now()}
}
```

- [ ] **Step 12: Run all tests**

Run: `go test ./core/shared/... -v`
Expected: ALL PASS

- [ ] **Step 13: Commit**

```bash
git add core/shared/domainevent/
git commit -m "refactor(shared): add base domain event contract"
```

---

## Task 2: `skillcatalog` — Value Objects, SkillCategory Aggregate, Published Language

**Files:**
- Create: `core/skillcatalog/domain/value_objects.go`
- Create: `core/skillcatalog/domain/value_objects_test.go`
- Create: `core/skillcatalog/domain/category.go`
- Create: `core/skillcatalog/domain/category_test.go`
- Create: `core/skillcatalog/domain/published.go`
- Modify: `core/skillcatalog/domain/installed_skill.go` (add accessor methods only)

### 2.1 Value objects

- [ ] **Step 1: Write failing tests for value objects**

```go
// core/skillcatalog/domain/value_objects_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
)

func TestSkillIDIsEmpty(t *testing.T) {
	assert.True(t, domain.SkillID("").IsEmpty())
	assert.False(t, domain.SkillID("abc-123").IsEmpty())
}

func TestSkillNameValidation(t *testing.T) {
	name, err := domain.NewSkillName("my-skill")
	assert.NoError(t, err)
	assert.Equal(t, "my-skill", name.String())

	_, err = domain.NewSkillName("")
	assert.Error(t, err)
}

func TestSkillSourceRefDerivesLogicalKey(t *testing.T) {
	ref := domain.SkillSourceRef{
		SourceURL:     "https://github.com/user/repo",
		SourceSubPath: "skills/my-skill",
	}
	key := ref.LogicalKey()
	assert.False(t, key.IsEmpty())
	assert.Contains(t, key.String(), "git:")
}

func TestSkillSourceRefEmptyWhenManual(t *testing.T) {
	ref := domain.SkillSourceRef{}
	assert.True(t, ref.LogicalKey().IsEmpty())
}

func TestSkillVersionStateHasUpdate(t *testing.T) {
	v := domain.SkillVersionState{SourceSHA: "abc", LatestSHA: "def"}
	assert.True(t, v.HasUpdate())

	v2 := domain.SkillVersionState{SourceSHA: "abc", LatestSHA: "abc"}
	assert.False(t, v2.HasUpdate())

	v3 := domain.SkillVersionState{SourceSHA: "abc", LatestSHA: ""}
	assert.False(t, v3.HasUpdate())
}

func TestSkillStorageRefIsEmpty(t *testing.T) {
	assert.True(t, domain.SkillStorageRef("").IsEmpty())
	assert.False(t, domain.SkillStorageRef("/some/path").IsEmpty())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/skillcatalog/domain/ -run TestSkillID -v`
Expected: FAIL

- [ ] **Step 3: Implement value objects**

```go
// core/skillcatalog/domain/value_objects.go
package domain

import (
	"errors"
	"strings"

	"github.com/shinerio/skillflow/core/shared/logicalkey"
)

var ErrInvalidSkillName = errors.New("invalid skill name")

// SkillID is the instance identity of an installed skill, local to skillcatalog.
type SkillID string

func (id SkillID) String() string { return string(id) }
func (id SkillID) IsEmpty() bool  { return strings.TrimSpace(string(id)) == "" }

// SkillName is a validated skill display name.
type SkillName string

func NewSkillName(raw string) (SkillName, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidSkillName
	}
	return SkillName(trimmed), nil
}

func (n SkillName) String() string { return string(n) }

// SkillSourceRef groups the fields that identify a skill's external source.
type SkillSourceRef struct {
	SourceURL     string
	SourceSubPath string
}

func (r SkillSourceRef) IsEmpty() bool {
	return strings.TrimSpace(r.SourceURL) == "" && strings.TrimSpace(r.SourceSubPath) == ""
}

func (r SkillSourceRef) LogicalKey() logicalkey.LogicalSkillKey {
	if strings.TrimSpace(r.SourceURL) == "" {
		return ""
	}
	key, err := logicalkey.GitFromRepoURL(r.SourceURL, r.SourceSubPath)
	if err != nil {
		return ""
	}
	return key
}

// SkillVersionState tracks source vs installed version for update detection.
type SkillVersionState struct {
	SourceSHA string
	LatestSHA string
}

func (v SkillVersionState) HasUpdate() bool {
	return v.LatestSHA != "" && v.LatestSHA != v.SourceSHA
}

// SkillStorageRef is the filesystem path where a skill is stored.
type SkillStorageRef string

func (r SkillStorageRef) String() string { return string(r) }
func (r SkillStorageRef) IsEmpty() bool  { return strings.TrimSpace(string(r)) == "" }
```

- [ ] **Step 4: Run tests**

Run: `go test ./core/skillcatalog/domain/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/skillcatalog/domain/value_objects.go core/skillcatalog/domain/value_objects_test.go
git commit -m "refactor(skillcatalog): add SkillID, SkillName, SkillSourceRef, SkillVersionState, SkillStorageRef value objects"
```

### 2.2 Add accessor methods to InstalledSkill (non-breaking bridge)

Instead of changing the struct field types (which would break 15+ files), add typed accessor methods on the existing `InstalledSkill` struct. This provides the DDD typed API while keeping the struct fields as `string` for now.

- [ ] **Step 6: Write failing test**

```go
// core/skillcatalog/domain/installed_skill_test.go (add to existing file)
func TestInstalledSkillAccessors(t *testing.T) {
	s := domain.InstalledSkill{
		ID:            "test-id",
		Name:          "my-skill",
		Path:          "/skills/my-skill",
		Source:        domain.SourceGitHub,
		SourceURL:     "https://github.com/user/repo",
		SourceSubPath: "skills/my-skill",
		SourceSHA:     "abc123",
		LatestSHA:     "def456",
	}
	assert.Equal(t, domain.SkillID("test-id"), s.SkillID())
	assert.Equal(t, domain.SkillName("my-skill"), s.SkillName())
	assert.Equal(t, domain.SkillStorageRef("/skills/my-skill"), s.StorageRef())

	ref := s.SourceRef()
	assert.Equal(t, "https://github.com/user/repo", ref.SourceURL)
	assert.Equal(t, "skills/my-skill", ref.SourceSubPath)
	assert.False(t, ref.LogicalKey().IsEmpty())

	vs := s.Version()
	assert.True(t, vs.HasUpdate())
}
```

- [ ] **Step 7: Run test to verify it fails**

Run: `go test ./core/skillcatalog/domain/ -run TestInstalledSkillAccessors -v`
Expected: FAIL

- [ ] **Step 8: Add accessor methods to InstalledSkill**

In `core/skillcatalog/domain/installed_skill.go`, add:

```go
func (s *InstalledSkill) SkillID() SkillID             { return SkillID(s.ID) }
func (s *InstalledSkill) SkillName() SkillName          { return SkillName(s.Name) }
func (s *InstalledSkill) StorageRef() SkillStorageRef    { return SkillStorageRef(s.Path) }

func (s *InstalledSkill) SourceRef() SkillSourceRef {
	return SkillSourceRef{SourceURL: s.SourceURL, SourceSubPath: s.SourceSubPath}
}

func (s *InstalledSkill) Version() SkillVersionState {
	return SkillVersionState{SourceSHA: s.SourceSHA, LatestSHA: s.LatestSHA}
}
```

- [ ] **Step 9: Run all tests**

Run: `go test ./core/... ./cmd/skillflow`
Expected: ALL PASS

- [ ] **Step 10: Commit**

```bash
git add core/skillcatalog/domain/installed_skill.go core/skillcatalog/domain/installed_skill_test.go
git commit -m "refactor(skillcatalog): add typed accessor methods on InstalledSkill"
```

### 2.3 SkillCategory aggregate root

Categories in SkillFlow are directory-based groupings. As an aggregate root, `SkillCategory` uses its `name` as identity (since category names are unique within the skill library).

- [ ] **Step 11: Write failing test**

```go
// core/skillcatalog/domain/category_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSkillCategory(t *testing.T) {
	cat, err := domain.NewSkillCategory("coding")
	assert.NoError(t, err)
	assert.Equal(t, "coding", cat.Name())
}

func TestSkillCategoryRejectsEmpty(t *testing.T) {
	_, err := domain.NewSkillCategory("")
	assert.Error(t, err)
}

func TestSkillCategoryRejectsInvalidChars(t *testing.T) {
	_, err := domain.NewSkillCategory("my/category")
	assert.Error(t, err)
}

func TestSkillCategoryEquality(t *testing.T) {
	a, _ := domain.NewSkillCategory("coding")
	b, _ := domain.NewSkillCategory("coding")
	assert.Equal(t, a.Name(), b.Name())
}
```

- [ ] **Step 12: Implement SkillCategory**

```go
// core/skillcatalog/domain/category.go
package domain

import (
	"errors"
	"strings"
)

var ErrInvalidCategoryName = errors.New("invalid category name")

// SkillCategory is an aggregate root representing a named skill grouping.
// Identity: the category name is unique within the skill library.
type SkillCategory struct {
	name string
}

func NewSkillCategory(raw string) (SkillCategory, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return SkillCategory{}, ErrInvalidCategoryName
	}
	if strings.ContainsAny(trimmed, `/\<>:"|?*`) {
		return SkillCategory{}, ErrInvalidCategoryName
	}
	return SkillCategory{name: trimmed}, nil
}

func (c SkillCategory) Name() string { return c.name }
```

- [ ] **Step 13: Run tests and commit**

Run: `go test ./core/skillcatalog/domain/ -v`

```bash
git add core/skillcatalog/domain/category.go core/skillcatalog/domain/category_test.go
git commit -m "refactor(skillcatalog): add SkillCategory aggregate root"
```

### 2.4 Published language types

- [ ] **Step 14: Create published types**

```go
// core/skillcatalog/domain/published.go
package domain

// InstalledSkillSummary is the published read-only view of an installed skill
// for cross-context consumption.
type InstalledSkillSummary struct {
	ID        SkillID
	Name      SkillName
	Category  string
	Source    SourceType
	SourceURL string
	Updatable bool
}

// InstalledSkillVersionView provides version comparison data for UI.
type InstalledSkillVersionView struct {
	ID        SkillID
	SourceSHA string
	LatestSHA string
	Updatable bool
}

// SkillCategorySummary is the published view of a category with its skill count.
type SkillCategorySummary struct {
	Name       string
	SkillCount int
}

// ToSummary creates a published summary from the aggregate.
func (s *InstalledSkill) ToSummary() InstalledSkillSummary {
	return InstalledSkillSummary{
		ID:        s.SkillID(),
		Name:      s.SkillName(),
		Category:  s.Category,
		Source:    s.Source,
		SourceURL: s.SourceURL,
		Updatable: s.HasUpdate(),
	}
}

// ToVersionView creates a version view from the aggregate.
func (s *InstalledSkill) ToVersionView() InstalledSkillVersionView {
	vs := s.Version()
	return InstalledSkillVersionView{
		ID:        s.SkillID(),
		SourceSHA: vs.SourceSHA,
		LatestSHA: vs.LatestSHA,
		Updatable: vs.HasUpdate(),
	}
}
```

- [ ] **Step 15: Run tests and commit**

Run: `go test ./core/skillcatalog/... -v`

```bash
git add core/skillcatalog/domain/published.go
git commit -m "refactor(skillcatalog): add InstalledSkillSummary, InstalledSkillVersionView, SkillCategorySummary published types"
```

---

## Task 3: `promptcatalog` — Value Objects, PromptCategory Aggregate, Published Language

**Files:**
- Create: `core/promptcatalog/domain/value_objects.go`
- Create: `core/promptcatalog/domain/value_objects_test.go`
- Create: `core/promptcatalog/domain/category.go`
- Create: `core/promptcatalog/domain/category_test.go`
- Create: `core/promptcatalog/domain/published.go`

### 3.1 Value objects

- [ ] **Step 1: Write failing tests**

```go
// core/promptcatalog/domain/value_objects_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/promptcatalog/domain"
	"github.com/stretchr/testify/assert"
)

func TestPromptIDIsEmpty(t *testing.T) {
	assert.True(t, domain.PromptID("").IsEmpty())
	assert.False(t, domain.PromptID("p-123").IsEmpty())
}

func TestPromptNameValidation(t *testing.T) {
	name, err := domain.NewPromptName("my-prompt")
	assert.NoError(t, err)
	assert.Equal(t, "my-prompt", name.String())

	_, err = domain.NewPromptName("")
	assert.Error(t, err)
}

func TestPromptContentIsEmpty(t *testing.T) {
	assert.True(t, domain.PromptContent("").IsEmpty())
	assert.False(t, domain.PromptContent("hello").IsEmpty())
}

func TestPromptStorageRefIsEmpty(t *testing.T) {
	assert.True(t, domain.PromptStorageRef("").IsEmpty())
	assert.False(t, domain.PromptStorageRef("/path").IsEmpty())
}

func TestPromptLinkSetLen(t *testing.T) {
	set := domain.PromptLinkSet{
		{Label: "a", URL: "https://a.com"},
		{Label: "b", URL: "https://b.com"},
	}
	assert.Equal(t, 2, set.Len())
	assert.Equal(t, 0, domain.PromptLinkSet(nil).Len())
}

func TestPromptMediaSetLen(t *testing.T) {
	set := domain.PromptMediaSet{"https://img1.png", "https://img2.png"}
	assert.Equal(t, 2, set.Len())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/promptcatalog/domain/ -run TestPromptID -v`
Expected: FAIL

- [ ] **Step 3: Implement value objects**

```go
// core/promptcatalog/domain/value_objects.go
package domain

import "strings"

// PromptID is the instance identity of a prompt.
type PromptID string

func (id PromptID) String() string { return string(id) }
func (id PromptID) IsEmpty() bool  { return strings.TrimSpace(string(id)) == "" }

// PromptName is a validated prompt display name.
type PromptName string

func NewPromptName(raw string) (PromptName, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidName
	}
	return PromptName(trimmed), nil
}

func (n PromptName) String() string { return string(n) }

// PromptContent is the text content of a prompt.
type PromptContent string

func (c PromptContent) String() string { return string(c) }
func (c PromptContent) IsEmpty() bool  { return strings.TrimSpace(string(c)) == "" }

// PromptStorageRef is the filesystem path where a prompt is stored.
type PromptStorageRef string

func (r PromptStorageRef) String() string { return string(r) }
func (r PromptStorageRef) IsEmpty() bool  { return strings.TrimSpace(string(r)) == "" }

// PromptLinkSet is a collection of web links attached to a prompt.
type PromptLinkSet []PromptLink

func (s PromptLinkSet) Len() int { return len(s) }

// PromptMediaSet is a collection of image URLs attached to a prompt.
type PromptMediaSet []string

func (s PromptMediaSet) Len() int { return len(s) }
```

- [ ] **Step 4: Run tests and commit**

Run: `go test ./core/promptcatalog/domain/ -v`

```bash
git add core/promptcatalog/domain/value_objects.go core/promptcatalog/domain/value_objects_test.go
git commit -m "refactor(promptcatalog): add PromptID, PromptName, PromptContent, PromptStorageRef, PromptLinkSet, PromptMediaSet value objects"
```

### 3.2 PromptCategory aggregate root

- [ ] **Step 5: Write failing test**

```go
// core/promptcatalog/domain/category_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/promptcatalog/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewPromptCategory(t *testing.T) {
	cat, err := domain.NewPromptCategory("writing")
	assert.NoError(t, err)
	assert.Equal(t, "writing", cat.Name())
}

func TestPromptCategoryDefaultNormalization(t *testing.T) {
	cat, err := domain.NewPromptCategory("default")
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultCategoryName, cat.Name())
}

func TestPromptCategoryEmptyNormalizesToDefault(t *testing.T) {
	cat, err := domain.NewPromptCategory("")
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultCategoryName, cat.Name())
	assert.True(t, cat.IsDefault())
}
```

- [ ] **Step 6: Implement PromptCategory**

```go
// core/promptcatalog/domain/category.go
package domain

import "strings"

// PromptCategory is an aggregate root representing a named prompt grouping.
// Identity: the category name is unique within the prompt library.
// Empty input normalizes to DefaultCategoryName.
type PromptCategory struct {
	name string
}

func NewPromptCategory(raw string) (PromptCategory, error) {
	normalized, err := NormalizeCategoryName(raw)
	if err != nil {
		return PromptCategory{}, err
	}
	return PromptCategory{name: normalized}, nil
}

func (c PromptCategory) Name() string    { return c.name }
func (c PromptCategory) IsDefault() bool { return strings.EqualFold(c.name, DefaultCategoryName) }
```

- [ ] **Step 7: Run tests and commit**

Run: `go test ./core/promptcatalog/domain/ -v`

```bash
git add core/promptcatalog/domain/category.go core/promptcatalog/domain/category_test.go
git commit -m "refactor(promptcatalog): add PromptCategory aggregate root"
```

### 3.3 Published language and Prompt accessor methods

- [ ] **Step 8: Create published types and add accessor methods to Prompt**

```go
// core/promptcatalog/domain/published.go
package domain

// PromptSummary is the published read-only view of a prompt for cross-context consumption.
type PromptSummary struct {
	Name        string
	Category    string
	Description string
	HasContent  bool
	LinkCount   int
	MediaCount  int
}

// PromptCategorySummary is the published view of a category with its prompt count.
type PromptCategorySummary struct {
	Name        string
	PromptCount int
}

func (p *Prompt) ToSummary() PromptSummary {
	return PromptSummary{
		Name:        p.Name,
		Category:    p.Category,
		Description: p.Description,
		HasContent:  !PromptContent(p.Content).IsEmpty(),
		LinkCount:   PromptLinkSet(p.WebLinks).Len(),
		MediaCount:  PromptMediaSet(p.ImageURLs).Len(),
	}
}
```

- [ ] **Step 9: Run tests and commit**

Run: `go test ./core/promptcatalog/... -v`

```bash
git add core/promptcatalog/domain/published.go
git commit -m "refactor(promptcatalog): add PromptSummary, PromptCategorySummary published types"
```

---

## Task 4: `agentintegration` — Value Objects, AgentPushPolicy Aggregate, Published Language

**Files:**
- Create: `core/agentintegration/domain/value_objects.go`
- Create: `core/agentintegration/domain/value_objects_test.go`
- Create: `core/agentintegration/domain/push_policy.go`
- Create: `core/agentintegration/domain/push_policy_test.go`
- Create: `core/agentintegration/domain/published.go`

### 4.1 Value objects

- [ ] **Step 1: Write failing tests**

```go
// core/agentintegration/domain/value_objects_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	"github.com/stretchr/testify/assert"
)

func TestAgentIDIsEmpty(t *testing.T) {
	assert.True(t, domain.AgentID("").IsEmpty())
	assert.False(t, domain.AgentID("claude-code").IsEmpty())
}

func TestAgentNameString(t *testing.T) {
	name := domain.AgentName("claude-code")
	assert.Equal(t, "claude-code", name.String())
}

func TestAgentTypeConstants(t *testing.T) {
	assert.Equal(t, domain.AgentTypeBuiltIn, domain.AgentType("builtin"))
	assert.Equal(t, domain.AgentTypeCustom, domain.AgentType("custom"))
}

func TestScanDirectorySetIsEmpty(t *testing.T) {
	assert.True(t, domain.ScanDirectorySet(nil).IsEmpty())
	assert.False(t, domain.ScanDirectorySet([]string{"/a"}).IsEmpty())
}

func TestPushDirectoryIsEmpty(t *testing.T) {
	assert.True(t, domain.PushDirectory("").IsEmpty())
	assert.False(t, domain.PushDirectory("/skills").IsEmpty())
}

func TestAgentSkillRefLogicalKey(t *testing.T) {
	ref := domain.AgentSkillRef{
		Name:       "my-skill",
		Path:       "/skills/my-skill",
		LogicalKey: logicalkey.LogicalSkillKey("content:abc123"),
	}
	assert.False(t, ref.LogicalKey.IsEmpty())
}

func TestPullConflict(t *testing.T) {
	c := domain.PullConflict{
		SkillName:  "my-skill",
		AgentName:  "claude-code",
		SourcePath: "/agent/skills/my-skill",
		Reason:     "already installed",
	}
	assert.Equal(t, "my-skill", c.SkillName)
}

func TestAgentSkillObservation(t *testing.T) {
	obs := domain.AgentSkillObservation{
		Name:       "my-skill",
		Path:       "/skills/my-skill",
		LogicalKey: logicalkey.LogicalSkillKey("content:abc123"),
		SeenInScan: true,
		Pushed:     false,
	}
	assert.True(t, obs.SeenInScan)
	assert.False(t, obs.LogicalKey.IsEmpty())
}
```

- [ ] **Step 2: Implement value objects**

```go
// core/agentintegration/domain/value_objects.go
package domain

import (
	"strings"

	"github.com/shinerio/skillflow/core/shared/logicalkey"
)

// AgentID is the unique identifier for an agent profile.
type AgentID string

func (id AgentID) String() string { return string(id) }
func (id AgentID) IsEmpty() bool  { return strings.TrimSpace(string(id)) == "" }

// AgentName is the display name of an agent.
type AgentName string

func (n AgentName) String() string { return string(n) }

// AgentType distinguishes built-in from custom agent profiles.
type AgentType string

const (
	AgentTypeBuiltIn AgentType = "builtin"
	AgentTypeCustom  AgentType = "custom"
)

// ScanDirectorySet is the list of directories an agent scans for skills.
type ScanDirectorySet []string

func (s ScanDirectorySet) IsEmpty() bool   { return len(s) == 0 }
func (s ScanDirectorySet) Paths() []string { return []string(s) }

// PushDirectory is the directory an agent receives pushed skills into.
type PushDirectory string

func (d PushDirectory) String() string { return string(d) }
func (d PushDirectory) IsEmpty() bool  { return strings.TrimSpace(string(d)) == "" }

// AgentSkillRef is a reference to a skill within an agent's directory.
type AgentSkillRef struct {
	Name       string
	Path       string
	LogicalKey logicalkey.LogicalSkillKey
}

// PullConflict represents a conflict detected when pulling a skill from an agent.
type PullConflict struct {
	SkillName  string
	AgentName  string
	SourcePath string
	Reason     string
}

// AgentSkillObservation records the presence of a skill in an agent's directories.
type AgentSkillObservation struct {
	Name       string
	Path       string
	LogicalKey logicalkey.LogicalSkillKey
	SeenInScan bool
	Pushed     bool
}
```

- [ ] **Step 3: Run tests and commit**

Run: `go test ./core/agentintegration/domain/ -v`

```bash
git add core/agentintegration/domain/value_objects.go core/agentintegration/domain/value_objects_test.go
git commit -m "refactor(agentintegration): add AgentID, AgentName, AgentType, ScanDirectorySet, PushDirectory, AgentSkillRef, PullConflict, AgentSkillObservation value objects"
```

### 4.2 AgentPushPolicy aggregate root

- [ ] **Step 4: Write failing test**

```go
// core/agentintegration/domain/push_policy_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
)

func TestAgentPushPolicyCanPush(t *testing.T) {
	policy := domain.AgentPushPolicy{
		AgentName: "claude-code",
		PushDir:   "/skills",
		Enabled:   true,
	}
	assert.True(t, policy.CanPush())
}

func TestAgentPushPolicyCannotPushWhenDisabled(t *testing.T) {
	policy := domain.AgentPushPolicy{
		AgentName: "claude-code",
		PushDir:   "/skills",
		Enabled:   false,
	}
	assert.False(t, policy.CanPush())
}

func TestAgentPushPolicyCannotPushWithoutDir(t *testing.T) {
	policy := domain.AgentPushPolicy{
		AgentName: "claude-code",
		PushDir:   "",
		Enabled:   true,
	}
	assert.False(t, policy.CanPush())
}
```

- [ ] **Step 5: Implement AgentPushPolicy**

```go
// core/agentintegration/domain/push_policy.go
package domain

import "strings"

// AgentPushPolicy is an aggregate root that captures the push rules for an agent.
type AgentPushPolicy struct {
	AgentName string
	PushDir   string
	Enabled   bool
}

func (p AgentPushPolicy) CanPush() bool {
	return p.Enabled && strings.TrimSpace(p.PushDir) != ""
}

func PushPolicyFromProfile(profile AgentProfile) AgentPushPolicy {
	return AgentPushPolicy{
		AgentName: profile.Name,
		PushDir:   profile.PushDir,
		Enabled:   profile.Enabled,
	}
}
```

- [ ] **Step 6: Run tests and commit**

Run: `go test ./core/agentintegration/domain/ -v`

```bash
git add core/agentintegration/domain/push_policy.go core/agentintegration/domain/push_policy_test.go
git commit -m "refactor(agentintegration): add AgentPushPolicy aggregate root"
```

### 4.3 Published language

- [ ] **Step 7: Create published types**

```go
// core/agentintegration/domain/published.go
package domain

import "github.com/shinerio/skillflow/core/shared/logicalkey"

// AgentSummary is the published read-only view of an agent for cross-context consumption.
type AgentSummary struct {
	Name    string
	Enabled bool
	HasPush bool
	HasScan bool
}

// AgentSkillPresence captures the resolved presence state of a skill across agents.
type AgentSkillPresence struct {
	LogicalKey   logicalkey.LogicalSkillKey
	Pushed       bool
	PushedAgents []string
}

// PushPlan describes a planned push operation for review before execution.
type PushPlan struct {
	AgentName  string
	TargetDir  string
	SkillNames []string
	Conflicts  []PushConflict
}

// PullPlan describes a planned pull operation for review before execution.
type PullPlan struct {
	AgentName  string
	SourceDir  string
	SkillNames []string
	Conflicts  []PullConflict
}

func (p AgentProfile) ToSummary() AgentSummary {
	return AgentSummary{
		Name:    p.Name,
		Enabled: p.Enabled,
		HasPush: PushPolicyFromProfile(p).CanPush(),
		HasScan: len(p.ScanDirs) > 0,
	}
}
```

- [ ] **Step 8: Run tests and commit**

Run: `go test ./core/agentintegration/... -v`

```bash
git add core/agentintegration/domain/published.go
git commit -m "refactor(agentintegration): add AgentSummary, AgentSkillPresence, PushPlan, PullPlan published types"
```

---

## Task 5: `skillsource` — Value Objects, SkillSource Aggregate, Published Language

**Files:**
- Create: `core/skillsource/domain/value_objects.go`
- Create: `core/skillsource/domain/value_objects_test.go`
- Create: `core/skillsource/domain/skill_source.go`
- Create: `core/skillsource/domain/skill_source_test.go`
- Create: `core/skillsource/domain/published.go`

### 5.1 Value objects

- [ ] **Step 1: Write failing tests**

```go
// core/skillsource/domain/value_objects_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
)

func TestStarRepoIDIsEmpty(t *testing.T) {
	assert.True(t, domain.StarRepoID("").IsEmpty())
	assert.False(t, domain.StarRepoID("repo-123").IsEmpty())
}

func TestSourceIDIsEmpty(t *testing.T) {
	assert.True(t, domain.SourceID("").IsEmpty())
}

func TestRepoSourceString(t *testing.T) {
	rs := domain.RepoSource("github.com/user/repo")
	assert.Equal(t, "github.com/user/repo", rs.String())
}

func TestSourceSubPathIsEmpty(t *testing.T) {
	assert.False(t, domain.SourceSubPath("skills/my-skill").IsEmpty())
	assert.True(t, domain.SourceSubPath("").IsEmpty())
}

func TestSourceSyncStatusValues(t *testing.T) {
	assert.Equal(t, domain.SyncStatusSynced, domain.SourceSyncStatus("synced"))
	assert.Equal(t, domain.SyncStatusError, domain.SourceSyncStatus("error"))
	assert.Equal(t, domain.SyncStatusPending, domain.SourceSyncStatus("pending"))
}

func TestSourceCacheRefIsEmpty(t *testing.T) {
	assert.True(t, domain.SourceCacheRef("").IsEmpty())
	assert.False(t, domain.SourceCacheRef("/cache/repo").IsEmpty())
}
```

- [ ] **Step 2: Implement value objects**

```go
// core/skillsource/domain/value_objects.go
package domain

import "strings"

// StarRepoID is the identity of a starred repository.
type StarRepoID string

func (id StarRepoID) String() string { return string(id) }
func (id StarRepoID) IsEmpty() bool  { return strings.TrimSpace(string(id)) == "" }

// SourceID is the identity of a skill source within a repository.
type SourceID string

func (id SourceID) String() string { return string(id) }
func (id SourceID) IsEmpty() bool  { return strings.TrimSpace(string(id)) == "" }

// RepoSource is the normalized repository identifier (e.g. "github.com/user/repo").
type RepoSource string

func (r RepoSource) String() string { return string(r) }

// SourceSubPath is the relative path within a repository that identifies a skill.
type SourceSubPath string

func (p SourceSubPath) String() string { return string(p) }
func (p SourceSubPath) IsEmpty() bool  { return strings.TrimSpace(string(p)) == "" }

// SourceSyncStatus tracks the synchronization state of a repository.
type SourceSyncStatus string

const (
	SyncStatusPending SourceSyncStatus = "pending"
	SyncStatusSynced  SourceSyncStatus = "synced"
	SyncStatusError   SourceSyncStatus = "error"
)

// SourceCacheRef is the local filesystem path where a repository cache is stored.
type SourceCacheRef string

func (r SourceCacheRef) String() string { return string(r) }
func (r SourceCacheRef) IsEmpty() bool  { return strings.TrimSpace(string(r)) == "" }
```

- [ ] **Step 3: Run tests and commit**

Run: `go test ./core/skillsource/domain/ -v`

```bash
git add core/skillsource/domain/value_objects.go core/skillsource/domain/value_objects_test.go
git commit -m "refactor(skillsource): add StarRepoID, SourceID, RepoSource, SourceSubPath, SourceSyncStatus, SourceCacheRef value objects"
```

### 5.2 SkillSource aggregate root

- [ ] **Step 4: Write failing test**

```go
// core/skillsource/domain/skill_source_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSkillSource(t *testing.T) {
	src := domain.NewSkillSource("github.com/user/repo", "skills/my-skill")
	assert.Equal(t, "github.com/user/repo", src.RepoSource())
	assert.Equal(t, "skills/my-skill", src.SubPath())
	assert.False(t, src.LogicalKey().IsEmpty())
}

func TestSkillSourceEquality(t *testing.T) {
	a := domain.NewSkillSource("github.com/user/repo", "skills/my-skill")
	b := domain.NewSkillSource("github.com/user/repo", "skills/my-skill")
	assert.Equal(t, a.LogicalKey(), b.LogicalKey())
}
```

- [ ] **Step 5: Implement SkillSource**

```go
// core/skillsource/domain/skill_source.go
package domain

import "github.com/shinerio/skillflow/core/shared/logicalkey"

// SkillSource is an aggregate root representing one logical skill source
// identified by repo + subpath. One StarRepo may contain many SkillSource entries.
type SkillSource struct {
	repoSource string
	subPath    string
}

func NewSkillSource(repoSource, subPath string) SkillSource {
	return SkillSource{
		repoSource: repoSource,
		subPath:    subPath,
	}
}

func (s SkillSource) RepoSource() string { return s.repoSource }
func (s SkillSource) SubPath() string    { return s.subPath }

func (s SkillSource) LogicalKey() logicalkey.LogicalSkillKey {
	return logicalkey.Git(s.repoSource, s.subPath)
}
```

- [ ] **Step 6: Run tests and commit**

Run: `go test ./core/skillsource/domain/ -v`

```bash
git add core/skillsource/domain/skill_source.go core/skillsource/domain/skill_source_test.go
git commit -m "refactor(skillsource): add SkillSource aggregate root"
```

### 5.3 Published language

- [ ] **Step 7: Create published types**

```go
// core/skillsource/domain/published.go
package domain

import (
	"time"

	"github.com/shinerio/skillflow/core/shared/logicalkey"
)

// StarRepoSummary is the published read-only view of a starred repository.
type StarRepoSummary struct {
	URL        string
	Name       string
	Source     string
	LastSync   time.Time
	SyncError  string
	SkillCount int
}

// SkillSourceSummary is the published view of a logical skill source.
type SkillSourceSummary struct {
	RepoSource string
	SubPath    string
	LogicalKey logicalkey.LogicalSkillKey
}

// SourceSkillCandidateView is the published view of a discovered skill candidate
// enriched with installed and pushed status from other contexts.
type SourceSkillCandidateView struct {
	Name         string
	Path         string
	SubPath      string
	RepoURL      string
	RepoName     string
	Source       string
	LogicalKey   logicalkey.LogicalSkillKey
	Installed    bool
	Imported     bool
	Updatable    bool
	Pushed       bool
	PushedAgents []string
}

// SourceVersionHint provides version comparison hints from the source side.
type SourceVersionHint struct {
	LogicalKey logicalkey.LogicalSkillKey
	LatestSHA  string
}

func (r *StarRepo) ToSummary(skillCount int) StarRepoSummary {
	return StarRepoSummary{
		URL:        r.URL,
		Name:       r.Name,
		Source:     r.Source,
		LastSync:   r.LastSync,
		SyncError:  r.SyncError,
		SkillCount: skillCount,
	}
}
```

- [ ] **Step 8: Run tests and commit**

Run: `go test ./core/skillsource/... -v`

```bash
git add core/skillsource/domain/published.go
git commit -m "refactor(skillsource): add StarRepoSummary, SkillSourceSummary, SourceSkillCandidateView, SourceVersionHint published types"
```

---

## Task 6: `backup` — Value Objects, GitBackupProfile

**Files:**
- Create: `core/backup/domain/value_objects.go`
- Create: `core/backup/domain/value_objects_test.go`

**Important:** `RestoreConflict` and other types that sound like they might conflict with existing `types.go` definitions are NEW types — no existing type with these names exists in `backup/domain/types.go`. Only `GitConflictError` exists there (which is different from `RestoreConflict`).

### 6.1 Value objects and GitBackupProfile

- [ ] **Step 1: Write failing tests**

```go
// core/backup/domain/value_objects_test.go
package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/backup/domain"
	"github.com/stretchr/testify/assert"
)

func TestBackupTargetIsEmpty(t *testing.T) {
	assert.True(t, domain.BackupTarget("").IsEmpty())
	assert.False(t, domain.BackupTarget("/data").IsEmpty())
}

func TestBackupScopeValues(t *testing.T) {
	assert.Equal(t, domain.BackupScopeFull, domain.BackupScope("full"))
	assert.Equal(t, domain.BackupScopeIncremental, domain.BackupScope("incremental"))
}

func TestBackupSnapshotLen(t *testing.T) {
	snap := domain.BackupSnapshot(domain.Snapshot{
		"a.json": {Size: 100, Hash: "abc"},
	})
	assert.Equal(t, 1, snap.Len())
}

func TestBackupChangeSet(t *testing.T) {
	cs := domain.BackupChangeSet{
		Added:   []string{"a.json"},
		Removed: []string{"b.json"},
	}
	assert.Equal(t, 1, len(cs.Added))
	assert.Equal(t, 1, len(cs.Removed))
}

func TestRestorePlan(t *testing.T) {
	plan := domain.RestorePlan{
		Provider:  "git",
		TargetDir: "/data",
	}
	assert.Equal(t, "git", plan.Provider)
}

func TestRestoreConflict(t *testing.T) {
	c := domain.RestoreConflict{
		Path:   "config.json",
		Reason: "local changes exist",
	}
	assert.Equal(t, "config.json", c.Path)
}

func TestGitBackupProfileIsGitBased(t *testing.T) {
	bp := domain.BackupProfile{
		Provider:   domain.GitProviderName,
		AppDataDir: "/data",
	}
	gbp := domain.NewGitBackupProfile(bp)
	assert.True(t, gbp.IsGitBased())
}

func TestGitBackupProfileNonGit(t *testing.T) {
	bp := domain.BackupProfile{
		Provider:   "s3",
		AppDataDir: "/data",
	}
	gbp := domain.NewGitBackupProfile(bp)
	assert.False(t, gbp.IsGitBased())
}
```

- [ ] **Step 2: Implement value objects and GitBackupProfile**

```go
// core/backup/domain/value_objects.go
package domain

import "strings"

// BackupTarget is the local directory that is the subject of backup.
type BackupTarget string

func (t BackupTarget) String() string { return string(t) }
func (t BackupTarget) IsEmpty() bool  { return strings.TrimSpace(string(t)) == "" }

// BackupScope describes the extent of a backup operation.
type BackupScope string

const (
	BackupScopeFull        BackupScope = "full"
	BackupScopeIncremental BackupScope = "incremental"
)

// BackupSnapshot wraps the Snapshot type for domain clarity.
type BackupSnapshot Snapshot

func (s BackupSnapshot) Len() int { return len(s) }

// BackupChangeSet tracks what files were added, modified, or removed in a backup cycle.
type BackupChangeSet struct {
	Added    []string
	Modified []string
	Removed  []string
}

// RestorePlan describes what a restore operation will do before execution.
type RestorePlan struct {
	Provider  string
	TargetDir string
	Files     []RemoteFile
	Conflicts []RestoreConflict
}

// RestoreConflict represents a conflict detected during restore planning.
type RestoreConflict struct {
	Path   string
	Reason string
}

// GitBackupProfile is an aggregate root for git-based backup configuration.
// It wraps BackupProfile with git-specific domain logic.
type GitBackupProfile struct {
	Profile BackupProfile
}

func NewGitBackupProfile(profile BackupProfile) GitBackupProfile {
	return GitBackupProfile{Profile: profile}
}

func (g GitBackupProfile) IsGitBased() bool {
	return g.Profile.Provider == GitProviderName
}
```

Note: `GitBackupProfile` uses **composition** (field `Profile`), not embedding, to keep aggregate root boundaries explicit.

- [ ] **Step 3: Run tests and commit**

Run: `go test ./core/backup/domain/ -v`

```bash
git add core/backup/domain/value_objects.go core/backup/domain/value_objects_test.go
git commit -m "refactor(backup): add BackupTarget, BackupScope, BackupSnapshot, BackupChangeSet, RestorePlan, RestoreConflict, GitBackupProfile"
```

---

## Task 7: Final Verification and Doc Alignment

- [ ] **Step 1: Run full test suite**

Run: `go test ./core/... ./cmd/skillflow`
Expected: ALL PASS

- [ ] **Step 2: Verify the patch 22d5c2e is intact**

Run: `git log --oneline -5`
Expected: Commit 22d5c2e should still be in history, no rebase or amend

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: Clean build

- [ ] **Step 4: Commit if any final adjustments were needed**

---

## Phase 2 Follow-up (separate plan)

After this plan completes and types are proven stable:

1. **Migrate `InstalledSkill` struct fields** from `string` to named types (`SkillID`, `SkillName`, `SkillStorageRef`). This requires updating 15+ files including the persistence layer (`filesystem_storage.go`), test files, readmodel, orchestration, and transport adapters.

2. **Migrate `AgentProfile` struct fields** to use `AgentName`, `ScanDirectorySet`, `PushDirectory`.

3. **Migrate `Prompt` struct fields** to use `PromptName`, `PromptContent`, `PromptStorageRef`, `PromptLinkSet`, `PromptMediaSet`.

4. **Migrate `StarRepo` struct fields** to include `StarRepoID`.

5. **Wire published language types** into `core/readmodel/` composers as the authoritative cross-context DTOs.
