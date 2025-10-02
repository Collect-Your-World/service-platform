package db

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
	sqlTrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"

	"backend/service-platform/app/internal/config"
)

type DB struct {
	PrimaryDb *bun.DB
	ReplicaDb *bun.DB
}

func NewDB(cfg config.ApplicationConfig, logger *zap.Logger) (db *DB, err error) {
	priDb, err := setupDatabase(
		"primary",
		cfg.DatabaseConfig.PrimaryConnectionString(),
		cfg.DatabaseConfig.MaxDBConns,
		cfg.DatabaseConfig.MaxIdleConns,
		cfg.DatabaseConfig.MaxConnLifetime,
		cfg.DatabaseConfig.MaxConnIdleTime,
		logger,
	)
	if err != nil {
		return nil, err
	}
	replDb, err := setupDatabase(
		"replica",
		cfg.DatabaseConfig.ReplicaConnectionString(),
		cfg.DatabaseConfig.MaxDBConns,
		cfg.DatabaseConfig.MaxIdleConns,
		cfg.DatabaseConfig.MaxConnLifetime,
		cfg.DatabaseConfig.MaxConnIdleTime,
		logger,
	)
	if err != nil {
		// Attempt to close the primary DB if the replica setup fails
		_ = priDb.Close()
		return nil, err
	}

	return &DB{
		PrimaryDb: priDb,
		ReplicaDb: replDb,
	}, nil
}

func setupDatabase(connType, dns string, maxOpen, maxIdle, maxLifetime, maxIdleTime int, logger *zap.Logger) (*bun.DB, error) {
	pgConnector := pgdriver.NewConnector(
		pgdriver.WithDSN(dns),
		pgdriver.WithTimeout(30*time.Second),
	)
	sqlTrace.Register("pgdriver", pgConnector.Driver(),
		sqlTrace.WithServiceName("db-job-service"),
		sqlTrace.WithAnalytics(true),
	)
	dbConn := sqlTrace.OpenDB(pgConnector)
	dbConn.SetMaxOpenConns(maxOpen)
	dbConn.SetMaxIdleConns(maxIdle)
	dbConn.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	dbConn.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)
	db := bun.NewDB(dbConn, pgdialect.New(), bun.WithDiscardUnknownColumns())
	err := db.Ping()
	if err != nil {
		logger.Error("failed to ping database", zap.String("type", connType), zap.Error(err))
		// Attempt to close the connection if ping fails
		_ = db.Close()
		return nil, fmt.Errorf("pinging %s: %w", connType, err)
	}
	logger.Info(
		"successfully connected to database",
		zap.String("type", connType),
		zap.Int("maxOpen", maxOpen),
		zap.Int("maxIdle", maxIdle),
		zap.Int("maxLifetime", maxLifetime),
		zap.Int("maxIdleTime", maxIdleTime),
	)
	return db, nil
}

func (d *DB) Close() error {
	var errPrimary, errReplica error
	if d.PrimaryDb != nil {
		errPrimary = d.PrimaryDb.Close()
	}
	if d.ReplicaDb != nil {
		errReplica = d.ReplicaDb.Close()
	}

	if errPrimary != nil {
		if errReplica != nil {
			return fmt.Errorf("error closing primary DB: %w; error closing replica DB: %v", errPrimary, errReplica)
		}
		return fmt.Errorf("error closing primary DB: %w", errPrimary)
	}
	if errReplica != nil {
		return fmt.Errorf("error closing replica DB: %w", errReplica)
	}
	return nil
}

func (d *DB) PrimaryConn() *bun.DB {
	return d.PrimaryDb
}

func (d *DB) ReplicaConn() *bun.DB {
	return d.ReplicaDb
}
