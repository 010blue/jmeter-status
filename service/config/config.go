package config

import (
	"database/sql"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql" // for mysql
	"log"
	"time"
)

// Task : Task struct
type Task struct {
	ID         int       `json:"id"`
	WebsiteID  int       `json:"website_id"`
	ExecutedAt time.Time `json:"executed_at"`
	Count      int       `json:"count"`
	ErrCount   int       `json:"err_count"`
	ErrRate    float32   `json:"err_rate"`
	File       string    `json:"file"`
}

// Website : JMeter website
type Website struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	URL          string      `json:"url"`
	AuthUser     string      `toml:"auth_user"`     // for 401
	AuthPassword string      `toml:"auth_password"` // for 401
	Data         []Task      `json:"data"`
	Days         []DayStatus `json:"days"`
}

// DayStatus : status data of day
type DayStatus struct {
	Date     string  `json:"date"`
	Count    int     `json:"count"`
	ErrCount int     `json:"err_count"`
	ErrRate  float32 `json:"err_rate"`
}

// MysqlConfig struct
type MysqlConfig struct {
	DSN string
}

// TomlConfig struct
type TomlConfig struct {
	Title    string
	Mysql    MysqlConfig
	Websites []Website
	Datapath string
	Rows     int
}

var configFile = "config/config.toml"

//InitConfig : initialze config
func InitConfig() (tomlConfig *TomlConfig, err error) {
	_, configErr := toml.DecodeFile(configFile, &tomlConfig)
	if configErr != nil {
		err = configErr
	}
	// websites sync to db
	err = SyncWebsitesToDB(tomlConfig)
	return tomlConfig, err
}

// InitDB : get DB object
func InitDB(tomlConfig *TomlConfig) (db *sql.DB) {
	db, err := sql.Open("mysql", tomlConfig.Mysql.DSN)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// SyncWebsitesToDB : sync websites of config.toml to db
func SyncWebsitesToDB(tomlConfig *TomlConfig) (err error) {
	db := InitDB(tomlConfig)
	defer db.Close()

	for _, website := range tomlConfig.Websites {
		websiteDb := db.QueryRow("SELECT id FROM websites WHERE id=?", website.ID)
		var id int
		websiteDb.Scan(&id)
		if sql.ErrNoRows != nil || id == 0 {
			stmt, stmtErr := db.Prepare("INSERT INTO websites(id,name,url,auth_user,auth_password,created_at,updated_at) VALUES(?,?,?,?,?,?,?)")
			if stmtErr != nil {
				return stmtErr
			}
			stmt.Exec(website.ID, website.Name, website.URL, website.AuthUser, website.AuthPassword, time.Now().UTC(), time.Now().UTC())
		}
	}

	return nil
}

// SyncTaskToDB : store task to DB
func SyncTaskToDB(task *Task, tomlConfig *TomlConfig) (err error) {
	db := InitDB(tomlConfig)
	defer db.Close()

	if task != nil {
		// judge if row exists by file
		taskRow := db.QueryRow("SELECT id FROM tasks WHERE `file`=?", task.File)
		var id int
		taskRow.Scan(&id)
		if taskRow == nil || id == 0 {
			stmt, stmtErr := db.Prepare("INSERT INTO tasks(website_id,file,api_count,api_error_count,api_error_rate,executed_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)")
			if stmtErr != nil {
				return stmtErr
			}
			_, execErr := stmt.Exec(task.WebsiteID, task.File, task.Count, task.ErrCount, task.ErrRate, task.ExecutedAt, time.Now().UTC(), time.Now().UTC())
			if execErr != nil {
				log.Println(execErr)
				return execErr
			}
		}
	}
	return err
}
