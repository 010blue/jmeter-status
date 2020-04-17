package config

import (
	"database/sql"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql" // for mysql
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

// Status : Api status
type Status struct {
	ID              int       `json:"id"`
	TaskID          int       `json:"task_id"`
	WebsiteID       int       `json:"website_id"`
	Position        string    `json:"position"`
	URL             string    `json:"url"`
	Label           string    `json:"label"`
	Timestamp       time.Time `json:"timestamp"`
	Filename        string    `json:"filename"`
	Elapsed         string    `json:"elapsed"`
	Method          string    `json:"method"`
	ResponseCode    string    `json:"response_code"`
	ResponseMessage string    `json:"response_message"`
	Success         string    `json:"success"`
	FailureMessage  string    `json:"failure_message"`
}

// SaveToDB : save status data to db
func (status Status) SaveToDB(tomlConfig *TomlConfig) (err error) {
	db := InitDB(tomlConfig)
	defer db.Close()

	// judge if row exists by file
	statusRow := db.QueryRow("SELECT id FROM statuses WHERE `id`=?", status.ID)
	var id int
	statusRow.Scan(&id)
	if statusRow == nil || id == 0 {
		stmt, stmtErr := db.Prepare("INSERT INTO statuses(task_id,website_id,position,url,label,timestamp,filename,elapsed,method,response_code,response_message,success,failure_message,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		if stmtErr != nil {
			log.Println(stmtErr)
			return stmtErr
		}
		_, execErr := stmt.Exec(status.TaskID, status.WebsiteID, status.Position, status.URL, status.Label, status.Timestamp, status.Filename, status.Elapsed, status.Method, status.ResponseCode, status.ResponseMessage, status.Success, status.FailureMessage, time.Now().UTC(), time.Now().UTC())
		if execErr != nil {
			log.Println(execErr)
			return execErr
		}
	}
	return err
}

// MysqlConfig struct
type MysqlConfig struct {
	DSN string
}

// PagerdutyConfig struct
type PagerdutyConfig struct {
	AuthToken         string `toml:"auth_token"`
	ServiceID         string `toml:"service_id"`
	From              string `toml:"from"`
	NotificationTitle string `toml:"notification_title"`
}

// NotificationConfig struct
type NotificationConfig struct {
	ShouldNotifyErrorNum int `toml:"should_notify_error_num"`
	Pagerduty            PagerdutyConfig
}

// TomlConfig struct
type TomlConfig struct {
	Title        string
	Mysql        MysqlConfig
	Notification NotificationConfig
	Websites     []Website
	Datapath     string
	Rows         int
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
		} else {
			// update
			stmt, stmtErr := db.Prepare("UPDATE tasks SET api_count = ?,api_error_count = ?,api_error_rate = ?,updated_at = ? WHERE id = ?")
			if stmtErr != nil {
				return stmtErr
			}
			_, execErr := stmt.Exec(task.Count, task.ErrCount, task.ErrRate, time.Now().UTC(), id)
			if execErr != nil {
				log.Println(execErr)
				return execErr
			}
		}
	}
	return err
}
