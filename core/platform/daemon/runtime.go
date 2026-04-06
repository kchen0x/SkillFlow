package daemon

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/config"
	memorycatalogapp "github.com/shinerio/skillflow/core/memorycatalog/app"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	sourcerepo "github.com/shinerio/skillflow/core/skillsource/infra/repository"
)

type MemoryServicesFactory func(cfgService *config.Service, dataDir string) (*memorycatalogapp.MemoryService, *memorycatalogapp.PushService)

type Dependencies struct {
	RunUpgrade          func(dataDir string) error
	LoadConfig          func(dataDir string) (*config.Service, config.AppConfig, error)
	NewLogger           func(logDir, level string) (*logging.Logger, error)
	SyncLaunchAtLogin   func(enabled bool) error
	RegisterAdapters    func()
	RegisterProviders   func()
	BuiltinStarredRepos []string
	NewMemoryServices   MemoryServicesFactory
}

type Runtime struct {
	DataDir           string
	ConfigService     *config.Service
	ConfigSnapshot    config.AppConfig
	ConfigLoadErr     error
	LoggerInitErr     error
	LaunchAtLoginErr  error
	Hub               *eventbus.Hub
	Logger            *logging.Logger
	Storage           *skillcatalogapp.Service
	StarStorage       *sourcerepo.StarRepoStorage
	CacheDir          string
	ViewCache         *viewstate.Manager
	MemoryService     *memorycatalogapp.MemoryService
	MemoryPushService *memorycatalogapp.PushService
	startupOnce       sync.Once
	stopAutoSync      chan struct{}
}

type StartupTask struct {
	Name  string
	Delay time.Duration
	Run   func()
}

func NewRuntime(dataDir string, deps Dependencies) (*Runtime, error) {
	runUpgrade := deps.RunUpgrade
	if runUpgrade == nil {
		runUpgrade = func(string) error { return nil }
	}
	loadConfig := deps.LoadConfig
	if loadConfig == nil {
		loadConfig = func(dataDir string) (*config.Service, config.AppConfig, error) {
			svc := config.NewService(dataDir)
			cfg, err := svc.Load()
			return svc, cfg, err
		}
	}
	newLogger := deps.NewLogger
	if newLogger == nil {
		newLogger = logging.New
	}

	if err := runUpgrade(dataDir); err != nil {
		return nil, err
	}

	rt := &Runtime{
		DataDir:  dataDir,
		Hub:      eventbus.NewHub(),
		CacheDir: filepath.Join(dataDir, "cache"),
	}

	cfgService, cfg, err := loadConfig(dataDir)
	if cfgService == nil {
		cfgService = config.NewService(dataDir)
	}
	rt.ConfigService = cfgService
	rt.ConfigLoadErr = err
	if err != nil {
		cfg = config.DefaultConfig(dataDir)
	}
	rt.ConfigSnapshot = cfg

	logger, loggerErr := newLogger(filepath.Join(dataDir, "logs"), cfg.LogLevel)
	rt.Logger = logger
	rt.LoggerInitErr = loggerErr

	if deps.SyncLaunchAtLogin != nil {
		rt.LaunchAtLoginErr = deps.SyncLaunchAtLogin(cfg.LaunchAtLogin)
	}

	rt.Storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(appdata.SkillsDir(dataDir)))
	rt.StarStorage = sourcerepo.NewStarRepoStorageWithBuiltinsAndCacheDir(
		filepath.Join(dataDir, "star_repos.json"),
		deps.BuiltinStarredRepos,
		repoCacheDir(cfgService, dataDir),
	)
	rt.ViewCache = viewstate.NewManager(filepath.Join(rt.CacheDir, "viewstate"))
	if deps.NewMemoryServices != nil {
		rt.MemoryService, rt.MemoryPushService = deps.NewMemoryServices(cfgService, dataDir)
	}
	if deps.RegisterAdapters != nil {
		deps.RegisterAdapters()
	}
	if deps.RegisterProviders != nil {
		deps.RegisterProviders()
	}

	return rt, nil
}

func repoCacheDir(cfgService *config.Service, dataDir string) string {
	if cfgService == nil {
		return appdata.RepoCacheDir(dataDir)
	}
	repoCacheDir := cfgService.LoadLocalRuntimeConfig().RepoCacheDir
	if repoCacheDir == "" {
		return appdata.RepoCacheDir(dataDir)
	}
	return repoCacheDir
}

func (rt *Runtime) ScheduleStartupTasks(tasks []StartupTask, schedule func(StartupTask)) {
	if rt == nil {
		return
	}
	rt.startupOnce.Do(func() {
		for _, task := range tasks {
			schedule(task)
		}
	})
}

func (rt *Runtime) StartAutoSyncTimer(intervalMinutes int, autoSync func()) {
	if rt == nil {
		return
	}
	if rt.stopAutoSync != nil {
		close(rt.stopAutoSync)
		rt.stopAutoSync = nil
	}
	if intervalMinutes <= 0 {
		return
	}
	stop := make(chan struct{})
	rt.stopAutoSync = stop
	go func() {
		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if autoSync != nil {
					autoSync()
				}
			case <-stop:
				return
			}
		}
	}()
}
