package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Store is the persistence boundary that the filter package depends on.
// Implementing it lets the gateway run against any storage backend.
type Store interface {
	GetOrCreateFilterGateway() (*FilterGateway, error)
	UpdateFilterGateway(gw *FilterGateway) error

	ListFilterApps() ([]FilterApp, error)
	GetFilterAppByName(name string) (*FilterApp, error)
	CreateFilterApp(app *FilterApp) error
	UpdateFilterApp(app *FilterApp) error
	DeleteFilterApp(id uint) error

	ListFilterTemplates() ([]FilterTemplate, error)
	GetFilterTemplate(id uint) (*FilterTemplate, error)
	GetDefaultFilterTemplate() (*FilterTemplate, error)
	CreateFilterTemplate(t *FilterTemplate) error
	UpdateFilterTemplate(t *FilterTemplate) error
	DeleteFilterTemplate(id uint) error
}

// DB is the default GORM-backed Store implementation.
type DB struct {
	gdb *gorm.DB
}

// gormStdLogger routes GORM's logger lines to os.Stderr without timestamping
// each one (zerolog handles real log entries; this is just for the rare GORM
// warning we don't want swallowed).
type gormStdLogger struct{}

func (gormStdLogger) Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Open opens (or creates) a SQLite database at path and runs migrations.
func Open(path string) (*DB, error) {
	if path == "" {
		path = "data/nextonebotfilter.db"
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}
	g, err := gorm.Open(sqlite.Open(abs+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"), &gorm.Config{
		Logger: logger.New(
			gormStdLogger{},
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := g.AutoMigrate(&FilterGateway{}, &FilterTemplate{}, &FilterApp{}); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	db := &DB{gdb: g}
	if _, err := db.GetOrCreateFilterGateway(); err != nil {
		return nil, err
	}
	if _, err := db.GetDefaultFilterTemplate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) GetOrCreateFilterGateway() (*FilterGateway, error) {
	var gw FilterGateway
	err := d.gdb.First(&gw, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		gw = FilterGateway{ID: 1}
		if err := d.gdb.Create(&gw).Error; err != nil {
			return nil, err
		}
		return &gw, nil
	}
	return &gw, err
}

func (d *DB) UpdateFilterGateway(gw *FilterGateway) error {
	gw.ID = 1
	return d.gdb.Save(gw).Error
}

func (d *DB) ListFilterApps() ([]FilterApp, error) {
	var apps []FilterApp
	err := d.gdb.Order("sort_order asc, id asc").Find(&apps).Error
	return apps, err
}

func (d *DB) GetFilterAppByName(name string) (*FilterApp, error) {
	var app FilterApp
	err := d.gdb.Where("name = ?", name).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (d *DB) CreateFilterApp(app *FilterApp) error { return d.gdb.Create(app).Error }
func (d *DB) UpdateFilterApp(app *FilterApp) error { return d.gdb.Save(app).Error }
func (d *DB) DeleteFilterApp(id uint) error        { return d.gdb.Delete(&FilterApp{}, id).Error }

func (d *DB) ListFilterTemplates() ([]FilterTemplate, error) {
	var ts []FilterTemplate
	err := d.gdb.Order("id asc").Find(&ts).Error
	return ts, err
}

func (d *DB) GetFilterTemplate(id uint) (*FilterTemplate, error) {
	var t FilterTemplate
	err := d.gdb.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *DB) GetDefaultFilterTemplate() (*FilterTemplate, error) {
	var t FilterTemplate
	err := d.gdb.Where("name = ?", "default").First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		t = FilterTemplate{
			Name:                "default",
			Description:         "Built-in default template; supplies global ID-rule fallback.",
			Builtin:             true,
			UserIDRules:         `{"mode":"on","ids":[]}`,
			GroupIDRules:        `{"mode":"on","ids":[]}`,
			MessageRules:        `{"mode":"on","filters":[],"prefix":[],"prefix_replace":""}`,
			PrivateMessageRules: `{"mode":"default","filters":[],"prefix":[],"prefix_replace":""}`,
			GroupMessageRules:   `{"mode":"default","filters":[],"prefix":[],"prefix_replace":""}`,
		}
		if err := d.gdb.Create(&t).Error; err != nil {
			return nil, err
		}
		return &t, nil
	}
	return &t, err
}

func (d *DB) CreateFilterTemplate(t *FilterTemplate) error { return d.gdb.Create(t).Error }
func (d *DB) UpdateFilterTemplate(t *FilterTemplate) error { return d.gdb.Save(t).Error }

func (d *DB) DeleteFilterTemplate(id uint) error {
	var t FilterTemplate
	if err := d.gdb.First(&t, id).Error; err != nil {
		return err
	}
	if t.Builtin {
		return errors.New("cannot delete builtin template")
	}
	return d.gdb.Delete(&FilterTemplate{}, id).Error
}
