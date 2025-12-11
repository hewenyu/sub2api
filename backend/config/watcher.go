package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type Watcher struct {
	configPath string
	config     *Config
	validator  *Validator
	onChange   func(*Config)
	logger     *zap.Logger
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

func NewWatcher(configPath string, config *Config, validator *Validator, onChange func(*Config), logger *zap.Logger) *Watcher {
	return &Watcher{
		configPath: configPath,
		config:     config,
		validator:  validator,
		onChange:   onChange,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watch()
}

func (w *Watcher) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

func (w *Watcher) watch() {
	defer w.wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Error("Failed to create file watcher", zap.Error(err))
		return
	}
	defer func() { _ = watcher.Close() }()

	if err := watcher.Add(w.configPath); err != nil {
		w.logger.Error("Failed to watch config file", zap.Error(err))
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				w.logger.Info("Config file changed, reloading")
				if err := w.reload(); err != nil {
					w.logger.Error("Failed to reload config", zap.Error(err))
				} else {
					w.logger.Info("Config reloaded successfully")
				}
			}

		case err := <-watcher.Errors:
			w.logger.Error("File watcher error", zap.Error(err))

		case <-w.stopChan:
			return
		}
	}
}

func (w *Watcher) reload() error {
	newConfig, err := LoadConfig(w.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := w.validator.Validate(newConfig); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	w.mu.Lock()
	w.config = newConfig
	w.mu.Unlock()

	if w.onChange != nil {
		w.onChange(newConfig)
	}

	return nil
}

func (w *Watcher) ReloadOnSignal(sig os.Signal) {
	w.logger.Info("Received reload signal", zap.String("signal", sig.String()))
	if err := w.reload(); err != nil {
		w.logger.Error("Failed to reload config", zap.Error(err))
	}
}

func (w *Watcher) GetConfig() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}
