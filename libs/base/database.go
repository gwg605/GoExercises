package base

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	Sqlite3DefaultDatabasePath = "%" + VAR_APP_ROOT_FOLDER_NAME + "%databases/%" + VAR_APP_NAME + "%.sqlite3"
	ServerTypeSqlite3          = "sqlite3"
	ServerTypePostgreSql       = "pgx"
	ServerTypeOption           = "type"
	Sqlite3DatabaseFile        = "database_file"
	PostgresqlOptionUrl        = "url"
	PostgresqlOptionDatabase   = "database"
	PostgresqlOptionEndpoint   = "endpoint"
	PostgresqlOptionUsername   = "username"
	PostgresqlOptionPassword   = "password"
)

type DB struct {
	DB         *sql.DB
	MakeParams ParamsMaker
}

func GetServerType(dbConfig Opts) string {
	return dbConfig.GetString(ServerTypeOption, ServerTypeSqlite3)
}

func SqlLite3GetDatabaseFile(dbConfig Opts) string {
	defaultStorePath := Sqlite3DefaultDatabasePath
	return ExpandString(dbConfig.GetString(Sqlite3DatabaseFile, defaultStorePath), dbConfig)
}

func OpenDatabase(dbConfig Opts) (*DB, error) {
	serverType := GetServerType(dbConfig)
	var connectionString = ""
	var paramsMaker ParamsMaker
	switch serverType {
	case ServerTypeSqlite3:
		databaseFile := SqlLite3GetDatabaseFile(dbConfig)
		databaseFolder, _ := filepath.Split(databaseFile)
		err := os.MkdirAll(databaseFolder, 0744)
		if err != nil {
			return nil, fmt.Errorf("unable to create database path '%s'. Error=%v", databaseFile, err)
		}
		connectionString += databaseFile
		paramsMaker = func() DatabaseParams { return &SqliteParams{} }
	case ServerTypePostgreSql:
		connectionURLStr := dbConfig.GetString(PostgresqlOptionUrl, "postgres://")
		connectionURL, err := url.Parse(connectionURLStr)
		if err != nil {
			return nil, fmt.Errorf("unable to parse '%s' = %v", connectionURLStr, err)
		}
		pwd, _ := connectionURL.User.Password()
		username := ExpandString(dbConfig.GetString(PostgresqlOptionUsername, connectionURL.User.Username()), dbConfig)
		password := ExpandString(dbConfig.GetString(PostgresqlOptionPassword, pwd), dbConfig)
		if username != "" || password != "" {
			connectionURL.User = url.UserPassword(username, password)
		}
		connectionURL.Path = ExpandString(dbConfig.GetString(PostgresqlOptionDatabase, connectionURL.Path), dbConfig)
		connectionURL.Host = ExpandString(dbConfig.GetString(PostgresqlOptionEndpoint, connectionURL.Host), dbConfig)
		connectionString = connectionURL.String()
		paramsMaker = func() DatabaseParams { return &PostgreSqlParams{} }
	default:
		return nil, fmt.Errorf("unknown '%s' SQL server type. Supported 'sqlite3'", serverType)
	}
	log.Printf("OpenDatabase() - serverType=%s, connectionString=%s", serverType, connectionString)
	db, err := sql.Open(serverType, connectionString)
	if err != nil {
		return nil, err
	}

	return &DB{DB: db, MakeParams: paramsMaker}, nil
}

func TimeToDBTime(t time.Time) sql.NullInt64 {
	if t.Equal(NilTime) {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.UnixNano(), Valid: true}
}

func DBTimeToTime(t sql.NullInt64) time.Time {
	if t.Valid {
		return time.Unix(0, t.Int64)
	}
	return NilTime
}
