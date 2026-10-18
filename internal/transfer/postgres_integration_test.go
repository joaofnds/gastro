//go:build transfer_integration

package transfer_test

import (
	"astro/internal/transfer"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgres transfer integration", func() {
	var db *sql.DB
	var schema string

	BeforeEach(func() {
		dsn := os.Getenv("ASTRO_TRANSFER_TEST_POSTGRES_DSN")
		Expect(dsn).NotTo(BeEmpty(), "explicit integration runs require a disposable Postgres DSN")
		var err error
		db, err = sql.Open("postgres", dsn)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(db.Close()).To(Succeed()) })
		Expect(db.PingContext(context.Background())).To(Succeed())
		schema = "transfer_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		_, err = db.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			_, cleanupErr := db.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
			Expect(cleanupErr).NotTo(HaveOccurred())
		})
		seedPostgres(db, schema)
	})

	It("round trips all four tables through a new SQLite file and isolated return schema", func() {
		path := filepath.Join(GinkgoT().TempDir(), "roundtrip.db")
		forward, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})
		Expect(err).NotTo(HaveOccurred())
		returnSchema := newReturnSchema(db)
		var database string
		Expect(db.QueryRow("SELECT current_database()").Scan(&database)).To(Succeed())

		returned, err := transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{Postgres: db, SQLitePath: path, TargetSchema: returnSchema, TargetDatabase: database, SourceFrozen: true})

		Expect(err).NotTo(HaveOccurred())
		Expect(returned).To(Equal(forward))
		verified, err := transfer.Verify(context.Background(), transfer.VerifyOptions{Postgres: db, PostgresSchema: returnSchema, SQLitePath: path})
		Expect(err).NotTo(HaveOccurred())
		Expect(verified).To(Equal(forward))
		original, err := transfer.Verify(context.Background(), transfer.VerifyOptions{Postgres: db, PostgresSchema: schema, SQLitePath: path})
		Expect(err).NotTo(HaveOccurred())
		Expect(original).To(Equal(forward))
	})

	It("preserves exact source fields and returns writes accepted after migration", func() {
		path := filepath.Join(GinkgoT().TempDir(), "new-writes.db")
		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})
		Expect(err).NotTo(HaveOccurred())
		target, err := sql.Open("sqlite", path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(target.Close()).To(Succeed()) })
		var name, description, owner, parent string
		var created time.Time
		Expect(target.QueryRow("SELECT user_id,name FROM habits WHERE id=?", habitID).Scan(&owner, &name)).To(Succeed())
		Expect([]string{owner, name}).To(Equal([]string{userID, "déjà vu\n\"read\""}))
		Expect(target.QueryRow("SELECT habit_id,description,created_at FROM activities WHERE id=?", activityID).Scan(&parent, &description, &created)).To(Succeed())
		Expect([]string{parent, description}).To(Equal([]string{habitID, "\"quoted\"\nactivity 🪴"}))
		Expect(created.UTC()).To(Equal(time.Date(2020, 2, 29, 23, 59, 58, 123456000, time.UTC)))
		Expect(target.QueryRow("SELECT name FROM groups WHERE id=? AND user_id=?", groupID, userID).Scan(&name)).To(Succeed())
		Expect(name).To(Equal("learning\n読書"))
		newID := "99767d2e-57f8-41c5-b34f-7b845a084d63"
		_, err = target.Exec("INSERT INTO activities(id,habit_id,description,created_at) VALUES(?,?,?,?)", newID, habitID, "after migration 🌱", "2026-10-18 12:00:00.000001+00:00")
		Expect(err).NotTo(HaveOccurred())
		returnSchema := newReturnSchema(db)
		var database string
		Expect(db.QueryRow("SELECT current_database()").Scan(&database)).To(Succeed())

		_, err = transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{Postgres: db, SQLitePath: path, TargetSchema: returnSchema, TargetDatabase: database, SourceFrozen: true})

		Expect(err).NotTo(HaveOccurred())
		Expect(db.QueryRow("SELECT description,created_at FROM "+qualified(returnSchema, "activities")+" WHERE id=$1", newID).Scan(&description, &created)).To(Succeed())
		Expect(description).To(Equal("after migration 🌱"))
		Expect(created.UTC()).To(Equal(time.Date(2026, 10, 18, 12, 0, 0, 1000, time.UTC)))
		var sourceRows int
		Expect(db.QueryRow("SELECT count(*) FROM "+qualified(schema, "activities")+" WHERE id=$1", newID).Scan(&sourceRows)).To(Succeed())
		Expect(sourceRows).To(BeZero())
	})

	It("preserves a genuine retained WAL and publishes no main file", func() {
		path, previous := initializedSQLite()
		seedSQLite(previous)
		wal, err := os.ReadFile(path + "-wal")
		Expect(err).NotTo(HaveOccurred())
		Expect(wal).NotTo(BeEmpty())
		sharedMemory, err := os.ReadFile(path + "-shm")
		Expect(err).NotTo(HaveOccurred())
		Expect(previous.Close()).To(Succeed())
		Expect(os.Remove(path)).To(Succeed())
		Expect(os.WriteFile(path+"-wal", wal, 0600)).To(Succeed())
		Expect(os.WriteFile(path+"-shm", sharedMemory, 0600)).To(Succeed())

		_, err = transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrTargetExists))
		Expect(path).NotTo(BeAnExistingFile())
		Expect(os.ReadFile(path + "-wal")).To(Equal(wal))
		Expect(os.ReadFile(path + "-shm")).To(Equal(sharedMemory))
	})

	It("preserves a previously published file on a repeated transfer", func() {
		path := filepath.Join(GinkgoT().TempDir(), "published.db")
		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})
		Expect(err).NotTo(HaveOccurred())
		original, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		_, err = transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrTargetExists))
		Expect(os.ReadFile(path)).To(Equal(original))
	})

	DescribeTable("detects a changed scalar field", func(statement string) {
		path := filepath.Join(GinkgoT().TempDir(), "verified.db")
		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})
		Expect(err).NotTo(HaveOccurred())
		sqliteDB, err := sql.Open("sqlite", path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(sqliteDB.Close()).To(Succeed()) })
		_, err = sqliteDB.Exec(statement)
		Expect(err).NotTo(HaveOccurred())

		_, err = transfer.Verify(context.Background(), transfer.VerifyOptions{Postgres: db, PostgresSchema: schema, SQLitePath: path})

		Expect(err).To(MatchError(transfer.ErrMismatch))
	},
		Entry("habit name", "UPDATE habits SET name='changed'"),
		Entry("group name", "UPDATE groups SET name='changed'"),
		Entry("activity identifier", "UPDATE activities SET id='55767d2e-57f8-41c5-b34f-7b845a084d63' WHERE id='"+activityID+"'"),
		Entry("activity description", "UPDATE activities SET description='changed'"),
		Entry("activity timestamp microsecond", "UPDATE activities SET created_at='2020-02-29 23:59:58.123457+00:00'"),
		Entry("group user key", "UPDATE groups SET user_id='66b67b16-f8e7-4686-8c78-bc7f5a70ed1a' WHERE name='empty group'"),
		Entry("membership group key", "UPDATE groups_habits SET group_id='33767d2e-57f8-41c5-b34f-7b845a084d63'"),
	)

	It("does not publish a partial target when source references violate SQLite constraints", func() {
		_, err := db.Exec("ALTER TABLE " + qualified(schema, "activities") + " DROP CONSTRAINT activities_habit_id_fkey")
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("UPDATE " + qualified(schema, "activities") + " SET habit_id='99767d2e-57f8-41c5-b34f-7b845a084d63'")
		Expect(err).NotTo(HaveOccurred())
		path := filepath.Join(GinkgoT().TempDir(), "failed.db")

		_, err = transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrWrite))
		Expect(path).NotTo(BeAnExistingFile())
		stages, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".astro-transfer-*"))
		Expect(err).NotTo(HaveOccurred())
		Expect(stages).To(BeEmpty())
		var unchanged int
		Expect(db.QueryRow("SELECT count(*) FROM " + qualified(schema, "activities") + " WHERE habit_id='99767d2e-57f8-41c5-b34f-7b845a084d63'").Scan(&unchanged)).To(Succeed())
		Expect(unchanged).To(Equal(2))
	})

	It("refuses unexpected source columns before publishing a file", func() {
		_, err := db.Exec("ALTER TABLE " + qualified(schema, "habits") + " ADD COLUMN unexpected text")
		Expect(err).NotTo(HaveOccurred())
		path := filepath.Join(GinkgoT().TempDir(), "unsupported-schema.db")

		_, err = transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: schema, OutputPath: path, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrValidation))
		Expect(path).NotTo(BeAnExistingFile())
	})

	It("rolls back the entire return schema when Postgres rejects a text value", func() {
		path, sqliteDB := initializedSQLite()
		seedSQLite(sqliteDB)
		_, err := sqliteDB.Exec("UPDATE activities SET description = ?", "before\x00after")
		Expect(err).NotTo(HaveOccurred())
		returnSchema := newReturnSchema(db)
		var database string
		Expect(db.QueryRow("SELECT current_database()").Scan(&database)).To(Succeed())

		_, err = transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{Postgres: db, SQLitePath: path, TargetSchema: returnSchema, TargetDatabase: database, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrWrite))
		var exists bool
		Expect(db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname=$1)", returnSchema).Scan(&exists)).To(Succeed())
		Expect(exists).To(BeFalse())
	})

	It("refuses an existing return schema without changing its tables", func() {
		path, sqliteDB := initializedSQLite()
		seedSQLite(sqliteDB)
		returnSchema := newReturnSchema(db)
		_, err := db.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(returnSchema))
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("CREATE TABLE " + qualified(returnSchema, "sentinel") + " (value text NOT NULL)")
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("INSERT INTO " + qualified(returnSchema, "sentinel") + " VALUES ('preserved')")
		Expect(err).NotTo(HaveOccurred())
		var database string
		Expect(db.QueryRow("SELECT current_database()").Scan(&database)).To(Succeed())

		_, err = transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{Postgres: db, SQLitePath: path, TargetSchema: returnSchema, TargetDatabase: database, SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrTargetExists))
		var unchanged string
		Expect(db.QueryRow("SELECT value FROM " + qualified(returnSchema, "sentinel")).Scan(&unchanged)).To(Succeed())
		Expect(unchanged).To(Equal("preserved"))
	})

	It("requires the exact target database confirmation before creating a return schema", func() {
		path, sqliteDB := initializedSQLite()
		seedSQLite(sqliteDB)
		returnSchema := newReturnSchema(db)

		_, err := transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{Postgres: db, SQLitePath: path, TargetSchema: returnSchema, TargetDatabase: "not_the_connected_database", SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrUnsafeTarget))
		var exists bool
		Expect(db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname=$1)", returnSchema).Scan(&exists)).To(Succeed())
		Expect(exists).To(BeFalse())
	})
})

func newReturnSchema(db *sql.DB) string {
	GinkgoHelper()
	schema := "astro_return_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	DeferCleanup(func() {
		_, err := db.Exec("DROP SCHEMA IF EXISTS " + pq.QuoteIdentifier(schema) + " CASCADE")
		Expect(err).NotTo(HaveOccurred())
	})

	return schema
}

func qualified(schema, table string) string {
	return pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier(table)
}

func seedPostgres(db *sql.DB, schema string) {
	GinkgoHelper()
	statements := []string{
		"CREATE TABLE " + qualified(schema, "habits") + " (id uuid PRIMARY KEY, user_id uuid NOT NULL, name varchar NOT NULL, UNIQUE(id,user_id))",
		"CREATE TABLE " + qualified(schema, "groups") + " (id uuid NOT NULL, user_id uuid NOT NULL, name varchar NOT NULL, PRIMARY KEY(id,user_id))",
		"CREATE TABLE " + qualified(schema, "activities") + " (id uuid PRIMARY KEY, habit_id uuid NOT NULL REFERENCES " + qualified(schema, "habits") + " (id) ON DELETE CASCADE, description varchar NOT NULL DEFAULT '', created_at timestamp NOT NULL DEFAULT now())",
		"CREATE TABLE " + qualified(schema, "groups_habits") + " (group_id uuid NOT NULL, habit_id uuid NOT NULL, user_id uuid NOT NULL, PRIMARY KEY(group_id,habit_id), FOREIGN KEY(habit_id,user_id) REFERENCES " + qualified(schema, "habits") + " (id,user_id) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY(group_id,user_id) REFERENCES " + qualified(schema, "groups") + " (id,user_id) ON DELETE CASCADE ON UPDATE CASCADE)",
	}
	for _, statement := range statements {
		_, err := db.Exec(statement)
		Expect(err).NotTo(HaveOccurred())
	}

	_, err := db.Exec("INSERT INTO "+qualified(schema, "habits")+" VALUES ($1,$2,$3),($4,$5,$6)", habitID, userID, "déjà vu\n\"read\"", "66767d2e-57f8-41c5-b34f-7b845a084d63", "76b67b16-f8e7-4686-8c78-bc7f5a70ed1a", "other user")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO "+qualified(schema, "groups")+" VALUES ($1,$2,$3),($1,$4,$5),($6,$2,$7)", groupID, userID, "learning\n読書", "76b67b16-f8e7-4686-8c78-bc7f5a70ed1a", "same group id other user", "33767d2e-57f8-41c5-b34f-7b845a084d63", "empty group")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO "+qualified(schema, "activities")+" VALUES ($1,$2,$3,$4),($5,$2,$6,$7)", activityID, habitID, "\"quoted\"\nactivity 🪴", "2020-02-29 23:59:58.123456", "88767d2e-57f8-41c5-b34f-7b845a084d63", "", "2000-01-01 00:00:00.000001")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO "+qualified(schema, "groups_habits")+" VALUES ($1,$2,$3)", groupID, habitID, userID)
	Expect(err).NotTo(HaveOccurred())
}
