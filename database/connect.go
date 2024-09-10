package database

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"key-server/logger"
	"key-server/model"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB connect to db
func ConnectDB() {
	var err error
	//p := config.Config("DB_PORT")
	p := os.Getenv("DB_PORT")
	port, err := strconv.ParseUint(p, 10, 32)

	if err != nil {
		panic("failed to parse database port")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log := logger.GetForFile("startup-errors")
		log.Error("Error while connecting to database",
			zap.Error(err),
		)
		panic("failed to connect to database")
	}

	fmt.Println("Connection Opened to Database")

	// List of model structs
	models := []interface{}{
		&model.EncryptionKey{},
		// Add the rest of your model structs here (ModelC, ModelD, ..., ModelJ)
	}

	// List of your model structs
	// Loop through your models and apply AutoMigrate
	for _, individualModel := range models {
		if err := db.AutoMigrate(individualModel); err != nil {
			panic("Failed to create table: " + err.Error())
		}
	}

	fmt.Println("Database Migrated")

	DB = DbInstance{
		Db: db,
	}
}

//var Ctx = context.Background()

func CreateRedisClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       dbNo,
	})
	return rdb
}

func CloseRedisClient(client *redis.Client) {
	err := client.Close()
	if err != nil {
		fmt.Println("Failed to close Redis client.")
	}
}
