package testcontainers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"valerygordeev/go/exercises/common"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgresql "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func createObjectsTable(db *sql.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS Objects (
		ID varchar(32) NOT NULL,
		Ver bigint NOT NULL,
		Val json NOT NULL,
		PRIMARY KEY (ID)
	)
	`
	_, err := db.Exec(sql)
	//log.Printf("db.Exec()=%v,%v", res, err)
	if err != nil {
		return err
	}
	return nil
}

func loadObject(db *sql.DB, id string) (int64, []byte, error) {
	sql := "SELECT Ver, Val FROM Objects WHERE ID = $1"
	row := db.QueryRow(sql, id)
	err := row.Err()
	if err != nil {
		return 0, nil, err
	}
	var version int64
	var value []byte
	err = row.Scan(&version, &value)
	if err != nil {
		return 0, nil, common.ErrorNotFound
	}
	return version, value, nil
}

func saveObject(db *sql.DB, version int64, id string, val []byte) (int64, error) {
	newVersion := time.Now().UTC().UnixMicro()
	sql := "UPDATE Objects SET Ver = $1, Val = $2 WHERE ID = $3 AND Ver = $4"
	res, err := db.Exec(sql, newVersion, val, id, version)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rows == 1 {
		return newVersion, nil
	}

	sql = "INSERT INTO Objects(ID, Ver, Val) VALUES($1, $2, $3)"
	res, err = db.Exec(sql, id, newVersion, val)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, common.ErrorInvalidState
			} else {
				log.Printf("code=%s, message=%s", pgErr.Code, pgErr.Message)
			}
		}
		return 0, err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rows == 1 {
		return newVersion, nil
	}
	return 0, fmt.Errorf("Fatal error")
}

func TestPostgreSqlBasic(t *testing.T) {
	ctx := context.Background()
	logConsumer := testcontainers.StdoutLogConsumer{}
	postgresC, err := tcpostgresql.Run(context.Background(),
		"postgres:16-alpine",
		tcpostgresql.WithDatabase("test"),
		tcpostgresql.WithUsername("user"),
		tcpostgresql.WithPassword("password"),
		tcpostgresql.BasicWaitStrategies(),
		testcontainers.WithLogConsumers(&logConsumer),
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

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Connected to PostgreSQL!")

	err = createObjectsTable(db)
	if err != nil {
		t.Fatal(err)
	}

	ver, val, err := loadObject(db, "object#0")
	log.Printf("ver=%d, val=%s, err=%v", ver, string(val), err)
	newVersion, err := saveObject(db, 0, "object#0", []byte("{\"a\": \"value0\"}"))
	log.Printf("newVersion=%d,%v", newVersion, err)
	ver, val, err = loadObject(db, "object#0")
	log.Printf("ver=%d, val=%s, err=%v", ver, string(val), err)
	newVersion1, err := saveObject(db, 0, "object#0", []byte("{\"a\": \"changed0\"}"))
	log.Printf("newVersion1=%d,%v", newVersion1, err)
	ver, val, err = loadObject(db, "object#0")
	log.Printf("ver=%d, val=%s, err=%v", ver, string(val), err)
	newVersion2, err := saveObject(db, ver, "object#0", []byte("{\"a\": \"changed1\"}"))
	log.Printf("newVersion2=%d,%v", newVersion2, err)
	ver, val, err = loadObject(db, "object#0")
	log.Printf("ver=%d, val=%s, err=%v", ver, string(val), err)
}

type Object struct {
	U []string `json:"u"`
}

func loadPGStats(db *sql.DB, index int) {
	log.Printf("[%d] loadPGStats:", index)
	stats := db.Stats()
	log.Printf("[%d] STAT: InUse=%d, Idle=%d, Conn=%d, Wait=%d, WaitDur=%v",
		index, stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount, stats.WaitDuration)
	log.Printf("[%d] ------------------------------------", index)
	rows, err := db.Query("SELECT datname, pid, state, query, application_name, wait_event FROM pg_stat_activity")
	if err != nil {
		log.Printf("loadPGStats - db.Query()=%v", err)
		return
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("Wraning! loadPGStats() - rows.Close() return %v error", err)
		}
	}()

	var datname sql.NullString
	var pid int
	var state sql.NullString
	var query sql.NullString
	var application_name sql.NullString
	var wait_event sql.NullString
	var rowIndex int
	for rows.Next() {
		err = rows.Scan(&datname, &pid, &state, &query, &application_name, &wait_event)
		if err != nil {
			log.Printf("loadPGStats - rows.Scan()=%v", err)
			return
		}
		log.Printf("[%d] STAT: %d|%s|%d|%s|%s|%s|%s", index, rowIndex, datname.String, pid, state.String, query.String, application_name.String, wait_event.String)
		rowIndex++
	}
}

func BenchmarkConcurentUpdates(b *testing.B) {
	ctx := context.Background()
	//logConsumer := testcontainers.StdoutLogConsumer{}
	postgresC, err := tcpostgresql.Run(context.Background(),
		"postgres:16-alpine",
		tcpostgresql.WithDatabase("test"),
		tcpostgresql.WithUsername("user"),
		tcpostgresql.WithPassword("password"),
		tcpostgresql.BasicWaitStrategies(),
		//testcontainers.WithLogConsumers(&logConsumer),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := postgresC.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	connectionString, err := postgresC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		b.Fatal(err)
	}

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	err = createObjectsTable(db)
	if err != nil {
		b.Fatal(err)
	}

	const parallelClients = 32
	actualClients := min(parallelClients, b.N)

	db.SetMaxOpenConns(parallelClients)
	db.SetMaxIdleConns(parallelClients / 2)

	wg := sync.WaitGroup{}

	var finishedClients atomic.Int32
	wg.Add(1)
	go func() {
		var index int
		for finishedClients.Load() < int32(actualClients) {
			loadPGStats(db, index)
			time.Sleep(500 * time.Millisecond)
			index += 1
		}
		wg.Done()
	}()

	b.ResetTimer()

	var totalCases atomic.Int64
	var totalLoadFail atomic.Int64
	var totalUpdateOk atomic.Int64
	var totalUpdateInvalidState atomic.Int64
	var totalUpdateFail atomic.Int64
	for c := range actualClients {
		wg.Add(1)
		go func() {
			for i := range calcIterations(b.N, actualClients, c) {
				_ = i
				totalCases.Add(1)
				var object Object
				objectIdx := rand.IntN(100)
				objectID := fmt.Sprintf("object#%d", objectIdx)
				ver, val, err := loadObject(db, objectID)
				if err != nil {
					if !errors.Is(err, common.ErrorNotFound) {
						log.Printf("[%d] loadObject(db, '%s) failed. Error=%v", c, objectID, err)
						totalLoadFail.Add(1)
						continue
					}
				} else {
					//log.Printf("[%d] loadObject('%s')", c, objectID)
					err = json.Unmarshal([]byte(val), &object)
					if err != nil {
						totalLoadFail.Add(1)
						continue
					}
				}
				if len(object.U) > 9 {
					object.U = object.U[1:]
				}
				object.U = append(object.U, time.Now().UTC().Format(time.RFC3339Nano))
				objectData, err := json.Marshal(object)
				if err != nil {
					totalUpdateFail.Add(1)
					continue
				}
				ver, err = saveObject(db, ver, objectID, objectData)
				if err != nil {
					if !errors.Is(err, common.ErrorInvalidState) {
						log.Printf("[%d] saveObject(db, %d, '%s) failed. Error=%v", c, ver, objectID, err)
						totalUpdateFail.Add(1)
					} else {
						totalUpdateInvalidState.Add(1)
					}
					continue
				}
				//log.Printf("[%d] saveObject('%s')", c, objectID)
				totalUpdateOk.Add(1)
			}
			finishedClients.Add(1)
			wg.Done()
		}()
	}

	wg.Wait()
	log.Printf("actualClients=%d, totalCases=%d, totalLoadFail=%d, totalUpdateOk=%d, totalUpdateInvalidState=%d, totalUpdateFail=%d",
		actualClients, totalCases.Load(), totalLoadFail.Load(), totalUpdateOk.Load(), totalUpdateInvalidState.Load(), totalUpdateFail.Load())
}
