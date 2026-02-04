package database

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/messenger/backend/internal/models"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func Initialize() (*gorm.DB, error) {
    dbHost := getEnv("DB_HOST", "localhost")
    dbPort := getEnv("DB_PORT", "5432")
    dbName := getEnv("DB_NAME", "messenger")
    dbUser := getEnv("DB_USER", "messenger")
    dbPassword := getEnv("DB_PASSWORD", "changeme")

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
        dbHost, dbPort, dbUser, dbPassword, dbName,
    )

    logLevel := logger.Info
    if os.Getenv("APP_ENV") == "production" {
        logLevel = logger.Error
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        PrepareStmt: true, // Query optimization: cache prepared statements
    })

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }

    maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 20)
    maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
    connMaxLifetime := getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)

    sqlDB.SetMaxIdleConns(maxIdleConns)
    sqlDB.SetMaxOpenConns(maxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
        log.Printf("Warning: Failed to create uuid-ossp extension: %v", err)
    }

    if err := autoMigrate(db); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

    log.Println("Database migrations and auto-migration completed successfully")

    return db, nil
}

func autoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Chat{},
        &models.ChatMember{},
        &models.Message{},
        &models.Channel{},
        &models.ChannelSubscriber{},
        &models.Subscription{},
        &models.PaymentLog{},
        &models.Contact{},
        &models.BlockedUser{},
        &models.MediaFile{},
        &models.Call{},
        &models.CallSignal{},
        &models.UserDevice{},
        &models.UserOneTimeKey{},
        &models.WikiPage{},
        &models.WikiRevision{},
        &models.CodeSnippet{},
        &models.TempRole{},
        &models.RSSFeed{},
        &models.RSSItem{},
        &models.AuditLog{},
    )
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    valueStr := getEnv(key, "")
    if value, err := strconv.Atoi(valueStr); err == nil {
        return value
    }
    return defaultValue
}
