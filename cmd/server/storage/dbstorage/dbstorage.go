package dbstorage

import (
	"context"
	"database/sql"
	"log"
	"runtime"
	"time"

	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

type DBStorage struct {
	databasePath string
	db           *sql.DB
}

func New(databasePath string) *DBStorage {
	log.Println("DBStorage::New::info: started")
	var res = &DBStorage{
		databasePath: databasePath,
	}

	var err error
	res.db, err = sql.Open("postgres", databasePath)
	if err != nil {
		log.Panic("DBStorage::New::error: in db open:", err)
	}
	runtime.SetFinalizer(res, storageFinalizer)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var value bool
	row := res.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE  table_name = 'metrics');`)
	err = row.Scan(&value)
	if err != nil {
		log.Panic("DBStorage::New::error: in table check:", err)
	}
	if !value {
		_, err = res.db.Exec(`CREATE TABLE metrics (mytype text, myid text, myvalue double precision, delta bigint, uid text UNIQUE);`)
		if err != nil {
			log.Panic("DBStorage::New::error: in table creation:", err)
		}
	} else {
		tableIsOk := true
		for _, name := range []string{"mytype", "myid", "myvalue", "delta", "uid"} {
			var value bool
			row := res.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT column_name FROM information_schema.columns WHERE table_name='metrics' and column_name=$1);`, name)
			err = row.Scan(&value)
			if err != nil {
				log.Panic("DBStorage::New::error: in columns check:", err)
			}
			if !value {
				tableIsOk = false
				break
			}
		}

		if !tableIsOk {
			log.Println("DBStorage::New::info: table is wrong, drop and create")
			_, err = res.db.Exec(`DROP TABLE metrics;`)
			if err != nil {
				log.Panic("DBStorage::New::error: in table drop:", err)
			}
			_, err = res.db.Exec(`CREATE TABLE metrics (mytype text, myid text, myvalue double precision, delta bigint, uid text UNIQUE);`)
			if err != nil {
				log.Panic("DBStorage::New::error: in table creation:", err)
			}
		} else {
			log.Println("DBStorage::New::info: existing table is OK")
		}
	}
	return res
}

func (s *DBStorage) SetCounterMetrics(name string, val metrics.Counter) {
	log.Println("DBStorage::SetCounterMetrics::info: started")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, `INSERT INTO metrics(mytype, myid, delta, uid) VALUES ('counter', $1, $2, $3) ON CONFLICT (uid) DO UPDATE SET delta = $2;`, name, val, "counter"+name)
	if err != nil {
		log.Println("DBStorage::SetCounterMetrics::error: in ExecContext:", err)
	}
}

func (s *DBStorage) GetCounterMetrics(name string) (metrics.Counter, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var value int64
	row := s.db.QueryRowContext(ctx, "SELECT delta FROM metrics WHERE mytype='counter' AND myid=$1;", name)
	err := row.Scan(&value)
	if err != nil {
		log.Println("DBStorage::GetCounterMetrics::info: in QueryRowContext:", err)
		return 0, false
	}
	return metrics.Counter(value), true
}

func (s *DBStorage) SetGaugeMetrics(name string, val metrics.Gauge) {
	log.Println("DBStorage::SetGaugeMetrics::info: started")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, "INSERT INTO metrics(mytype, myid, myvalue, uid) VALUES ('gauge', $1, $2, $3) ON CONFLICT (uid) DO UPDATE SET myvalue = $2;", name, val, "gauge"+name)
	if err != nil {
		log.Println("DBStorage::SetGaugeMetrics::error: in ExecContext:", err)
	}
}

func (s *DBStorage) GetGaugeMetrics(name string) (metrics.Gauge, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var value float64
	row := s.db.QueryRowContext(ctx, "SELECT myvalue FROM metrics WHERE mytype='gauge' AND myid=$1;", name)
	err := row.Scan(&value)
	if err != nil {
		log.Println("DBStorage::GetGaugeMetrics::info: in QueryRowContext:", err)
		return 0, false
	}
	return metrics.Gauge(value), true
}

func (s *DBStorage) GetKnownMetrics() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var res []string
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT myid FROM metrics")
	if err != nil {
		log.Println("DBStorage::GetKnownMetrics::info: in QueryContext:", err)
		return res
	}
	if rows.Err() != nil {
		log.Println("DBStorage::GetKnownMetrics::error: in rows:", err)
		return res
	}
	for rows.Next() {
		var val string
		err := rows.Scan(&val)
		if err != nil {
			log.Println("DBStorage::GetKnownMetrics::error: in scan:", err)
			continue
		}
		res = append(res, val)
	}
	return res
}

func (s *DBStorage) IsDBConnected() bool {
	if s.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func storageFinalizer(s *DBStorage) {
	log.Println("DBStorage::storageFinalizer::info: started")
	defer s.db.Close()
}
