package db

import (
    "fmt"
    "time"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Config struct {
    Driver   string
    Host     string
    Port     string
    User     string
    Password string
    Database string
    SSLMode  string
    Debug    bool
}

func Connect(cfg *Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
    )

    logLevel := logger.Silent
    if cfg.Debug {
        logLevel = logger.Info
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
        NowFunc: func() time.Time { return time.Now().UTC() },
    })
    if err != nil {
        return nil, err
    }

    if cfg.Driver == "cockroachdb" {
        db.Exec("SET default_transaction_isolation = 'serializable'")
    }

    return db, nil
}

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Job{},
        &User{},
    )
}