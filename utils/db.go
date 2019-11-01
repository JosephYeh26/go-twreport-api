package utils

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jinzhu/gorm"
	"gopkg.in/matryer/try.v1"
	"gopkg.in/mgo.v2"

	"twreporter.org/go-api/globals"
	"twreporter.org/go-api/models"
)

// InitDB initiates the MySQL database connection
func InitDB(attempts, retryMaxDelay int) (*gorm.DB, error) {
	var db *gorm.DB
	err := try.Do(func(attempt int) (bool, error) {
		var err error
		var config = globals.Conf.DB.MySQL

		// connect to MySQL database
		var endpoint = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4,utf8&parseTime=true", config.User, config.Password, config.Address, config.Port, config.Name)
		log.Info("connect to mysql ", endpoint)
		db, err = gorm.Open("mysql", endpoint)

		if err != nil {
			time.Sleep(time.Duration(retryMaxDelay) * time.Second)
		}

		return attempt < attempts, err
	})

	if err != nil {
		log.Error("Please check the MySQL database connection: ", err.Error())
		return nil, err
	}

	db.SetJoinTableHandler(&models.User{}, globals.TableBookmarks, &models.UsersBookmarks{})

	//db.LogMode(true)

	return db, nil
}

// InitMongoDB initiates the Mongo DB connection
func InitMongoDB() (*mgo.Session, error) {
	var timeout = globals.Conf.DB.Mongo.Timeout
	// Set connection timeout
	session, err := mgo.DialWithTimeout(globals.Conf.DB.Mongo.URL, time.Duration(timeout)*time.Second)
	log.Info("connect to mongodb ", globals.Conf.DB.Mongo.URL)

	if err != nil {
		log.Error("Establishing a new session to the mongo occurs error: ", err.Error())
		return nil, err
	}

	// Set operation timeout
	session.SetSyncTimeout(time.Duration(timeout) * time.Second)

	// Set socket timeout to 3 mins
	session.SetSocketTimeout(3 * time.Minute)

	// Since we don't have writes here and don't care about the consistency between Mongo Master and Slave,
	// we choose Eventual mode here.
	// The Eventual mode is the fastest and most resource-friendly,
	// but is also the one offering the least guarantees about ordering of the data read and written.
	session.SetMode(mgo.Eventual, true)

	return session, nil
}

// Get the migrate instance for operating migration
func GetMigrateInstance(dbInstance *sql.DB) (*migrate.Migrate, error) {
	const migrateMysqlDriver = "mysql"
	const migrateSourceDriver = "file"
	var migrateSourceDir string = filepath.Join(GetProjectRoot(), "migrations")

	driver, _ := mysql.WithInstance(dbInstance, &mysql.Config{})

	sourceUrl := fmt.Sprintf("%s://%s", migrateSourceDriver, migrateSourceDir)
	return migrate.NewWithDatabaseInstance(sourceUrl, migrateMysqlDriver, driver)
}
