package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string) (*postgres, error) {
	// singleton
	pgOnce.Do(func() {
		dbpool, err := pgxpool.New(ctx, connString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			return
		}

		pgInstance = &postgres{dbpool}
	})

	return pgInstance, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

type calendarEvent struct {
	title       string
	description string
	startDate   string
	endDate     string
}

// INSERT
func (pg *postgres) InsertUser(ctx context.Context, evt calendarEvent) error {
	query := `INSERT INTO event (title, description, start_date, end_date) VALUES (@eventTitle, @eventDesc, @eventStartDate, @eventEndDate)`
	args := pgx.NamedArgs{
		"eventTitle":     evt.title,
		"eventDesc":      evt.description,
		"eventStartDate": evt.startDate,
		"eventEndDate":   evt.endDate,
	}

	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

// Batch INSERT
func (pg *postgres) BulkInsertUsers(ctx context.Context, evts []calendarEvent) error {
	query := `INSERT INTO event (title, description, start_date, end_date) VALUES (@eventTitle, @eventDesc, @eventStartDate, @eventEndDate)`

	batch := &pgx.Batch{}
	for _, e := range evts {
		args := pgx.NamedArgs{
			"eventTitle":     e.title,
			"eventDesc":      e.description,
			"eventStartDate": e.startDate,
			"eventEndDate":   e.endDate,
		}
		batch.Queue(query, args)
	}

	results := pg.db.SendBatch(ctx, batch)
	defer results.Close()

	for _, e := range evts {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				log.Printf("user %s already exists", e.title)
				continue
			}

			return fmt.Errorf("unable to insert row: %w", err)
		}
	}

	return results.Close()
}

// COPY
func (pg *postgres) CopyInsertEvents(ctx context.Context, evts []calendarEvent) error {
	entries := [][]any{}
	columns := []string{"name", "email"}
	tableName := "event"

	for _, e := range evts {
		entries = append(entries, []any{e.title, e.description, e.startDate, e.endDate})
	}

	_, err := pg.db.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)

	if err != nil {
		return fmt.Errorf("error copying into %s table: %w", tableName, err)
	}

	return nil
}

// GET
func (pg *postgres) GetEvent(ctx context.Context, name string) ([]calendarEvent, error) {
	query := `SELECT name, email FROM user LIMIT 10`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[calendarEvent])
}
