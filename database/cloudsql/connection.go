package cloudsql

import (
	"SureMFService/config"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s sslmode=%s",
		config.AppConfig.DBHost,
		config.AppConfig.DBUser,
		config.AppConfig.DBName,
		config.AppConfig.DBPort,
		config.AppConfig.DBPassword,
		config.AppConfig.DBSSLMode,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		Logger:                 newLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.AppConfig.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.AppConfig.DBConnMaxLifetime)

	log.Println("Connected to PostgreSQL (sure_mf schema)")
	return nil
}
