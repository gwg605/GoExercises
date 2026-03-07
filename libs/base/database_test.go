package base_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
	"valerygordeev/go/exercises/libs/base"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	_ "github.com/ncruces/go-sqlite3/vfs/memdb"

	_ "github.com/jackc/pgx/v5/stdlib"

	tcpostgresql "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestParams(t *testing.T) {
	postgresqlParams := base.PostgreSqlParams{}
	next := postgresqlParams.Next()
	if next != "$1" {
		t.Fatalf("Wrong Next() %s. Expected $1", next)
	}
	next = postgresqlParams.Next()
	if next != "$2" {
		t.Fatalf("Wrong Next() %s. Expected $2", next)
	}
	postgresqlParams.Reset()
	next = postgresqlParams.Next()
	if next != "$1" {
		t.Fatalf("Wrong Next() %s. Expected $1", next)
	}

	sqlite3Params := base.SqliteParams{}
	next = sqlite3Params.Next()
	if next != "?" {
		t.Fatalf("Wrong Next() %s. Expected ?", next)
	}
}

func TestUtils(t *testing.T) {
	tm := time.Now()
	dbTm := base.TimeToDBTime(tm)
	tmCheck := base.DBTimeToTime(dbTm)
	if !tm.Equal(tmCheck) {
		t.Fatalf("Wrong time %v. Expected %v", tmCheck, tm)
	}

	tm = base.NilTime
	dbTm = base.TimeToDBTime(tm)
	dbNull := sql.NullInt64{}
	if dbTm.Valid != dbNull.Valid || dbTm.Int64 != dbNull.Int64 {
		t.Fatalf("Wrong dbTm %v. Expected %v", dbTm, dbNull)
	}
	tmCheck = base.DBTimeToTime(dbTm)
	if !tm.Equal(tmCheck) {
		t.Fatalf("Wrong time %v. Expected %v", tmCheck, tm)
	}
}

func TestOpenDatabaseFails(t *testing.T) {
	db, err := base.OpenDatabase(base.Opts{base.ServerTypeOption: "wrong-type"})
	if err == nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return error")
	}
	if db != nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return nil for DB")
	}

	db, err = base.OpenDatabase(base.Opts{base.ServerTypeOption: base.ServerTypeSqlite3, base.Sqlite3DatabaseFile: "/lost+found/not.found/none"})
	if err == nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return error")
	}
	if db != nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return nil for DB")
	}

	db, err = base.OpenDatabase(base.Opts{base.ServerTypeOption: base.ServerTypePostgreSql, base.PostgresqlOptionUrl: "\r"})
	if err == nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return error")
	}
	if db != nil {
		t.Errorf("base.OpenDatabase(wrong-type) must return nil for DB")
	}
}

func BasicTest(t *testing.T, db *base.DB) {
	sqlCreateTable := `
	CREATE TABLE table0 (
		ID int NOT NULL,
		Ver int NOT NULL,
		Val TEXT NOT NULL,
		PRIMARY KEY (ID)
	)
	`
	_, err := db.DB.Exec(sqlCreateTable)
	if err != nil {
		t.Fatalf("db.Exec(sqlCreateTable)=%v", err)
	}

	dbParams := db.MakeParams()
	sqlInsert := fmt.Sprintf("INSERT INTO table0 (ID, Ver, Val) VALUES(%s, %s, %s)", dbParams.Next(), dbParams.Next(), dbParams.Next())
	res, err := db.DB.Exec(sqlInsert, 1, 1, "Value#1")
	if err != nil {
		t.Fatalf("db.Exec(sqlInsert)=%v", err)
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("res.RowsAffected()=%v", err)
	}
	if rowsCount != 1 {
		t.Fatalf("Wrong rowsCount %d, Expected %d", rowsCount, 1)
	}

	dbParams.Reset()
	sqlSelect := fmt.Sprintf("SELECT ID, Ver, Val FROM table0 WHERE ID = %s", dbParams.Next())
	rows, err := db.DB.Query(sqlSelect, 1)
	if err != nil {
		t.Fatalf("db.Query(sqlSelect)=%v", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("rows.Close()=%v", err)
		}
	}()
	var selectedRowsCount int
	var id int64
	var ver int64
	var val string
	for rows.Next() {
		err = rows.Scan(&id, &ver, &val)
		if err != nil {
			t.Fatalf("rows.Scan()=%v", err)
		}
		if id != 1 {
			t.Fatalf("Wrong ID %d, expected 1", id)
		}
		if ver != 1 {
			t.Fatalf("Wrong Ver %d, expected 1", ver)
		}
		if val != "Value#1" {
			t.Fatalf("Wrong Ver %s, expected 'Value#1'", val)
		}
		selectedRowsCount++
	}
	if selectedRowsCount != 1 {
		t.Fatalf("Wrong selectedRowsCount %d. Expected 1", selectedRowsCount)
	}
}

func TestSQLiteBasic(t *testing.T) {
	err := base.InitVars("", "test", false)
	if err != nil {
		t.Fatalf("base.InitVars()=%v", err)
	}

	dbFileName := base.SqlLite3GetDatabaseFile(base.Opts{})
	dbFileNameExpected := base.EnsureTrailingPathSeparator(filepath.Join(os.Getenv("HOME"), "/valerygordeev/test/databases/")) + "test.sqlite3"
	if dbFileName != dbFileNameExpected {
		t.Errorf("Wrong dbFileName '%s'. Expected '%s'", dbFileName, dbFileNameExpected)
	}

	tempDBFileName := base.TempFileName("test", ".sqlite3")
	db, err := base.OpenDatabase(base.Opts{base.ServerTypeOption: base.ServerTypeSqlite3, base.Sqlite3DatabaseFile: tempDBFileName})
	if err != nil {
		t.Fatalf("OpenDatabase()=%v", err)
	}
	defer func() {
		err := db.DB.Close()
		if err != nil {
			log.Printf("db.Close()=%v", err)
		}
		err = os.Remove(tempDBFileName)
		if err != nil {
			log.Printf("os.Remove(%s)=%v", tempDBFileName, err)
		}
	}()

	BasicTest(t, db)
}

func TestPostgreSqlBasic(t *testing.T) {
	ctx := context.Background()
	postgresC, err := tcpostgresql.Run(context.Background(),
		"postgres:16-alpine",
		tcpostgresql.WithDatabase("test"),
		tcpostgresql.WithUsername("user"),
		tcpostgresql.WithPassword("password"),
		tcpostgresql.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := postgresC.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	connectionString, err := postgresC.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("connectionString=%s", connectionString)

	db, err := base.OpenDatabase(base.Opts{base.ServerTypeOption: base.ServerTypePostgreSql, "url": connectionString})
	if err != nil {
		t.Fatalf("OpenDatabase()=%v", err)
	}
	defer func() {
		err := db.DB.Close()
		if err != nil {
			log.Printf("db.Close()=%v", err)
		}
	}()

	BasicTest(t, db)
}
