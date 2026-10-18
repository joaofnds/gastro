package transfer_test

import (
	astrosqlite "astro/adapters/sqlite"
	"astro/internal/transfer"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransfer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transfer")
}

var _ = Describe("ToSQLite", func() {
	DescribeTable("preserves retained destination files", func(name string) {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		retained := filepath.Join(filepath.Dir(path), name)
		original := []byte("retained database state")
		Expect(os.WriteFile(retained, original, 0600)).To(Succeed())

		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{OutputPath: path, SourceFrozen: true, PostgresSchema: "public"})

		Expect(err).To(MatchError(transfer.ErrTargetExists))
		Expect(os.ReadFile(retained)).To(Equal(original))
		Expect(path).NotTo(BeAnExistingFile())
	}, Entry("WAL", "astro.db-wal"), Entry("shared memory", "astro.db-shm"), Entry("replication state", ".astro.db-litestream"))

	It("refuses an existing target without touching its contents", func() {
		path := filepath.Join(GinkgoT().TempDir(), "existing.db")
		original := []byte("previously published database")
		Expect(os.WriteFile(path, original, 0600)).To(Succeed())

		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{OutputPath: path, SourceFrozen: true, PostgresSchema: "public"})

		Expect(err).To(MatchError(transfer.ErrTargetExists))
		Expect(os.ReadFile(path)).To(Equal(original))
	})

	It("requires the operator to freeze source writes", func() {
		path := filepath.Join(GinkgoT().TempDir(), "new.db")

		_, err := transfer.ToSQLite(context.Background(), transfer.ToSQLiteOptions{OutputPath: path, PostgresSchema: "public"})

		Expect(err).To(MatchError(transfer.ErrSourceNotFrozen))
		Expect(path).NotTo(BeAnExistingFile())
	})
})

var _ = Describe("ToPostgres", func() {
	DescribeTable("refuses schemas that could select existing production tables", func(schema string) {
		_, err := transfer.ToPostgres(context.Background(), transfer.ToPostgresOptions{TargetSchema: schema, TargetDatabase: "astro", SourceFrozen: true})

		Expect(err).To(MatchError(transfer.ErrUnsafeTarget))
	},
		Entry("public schema", "public"),
		Entry("empty schema", ""),
		Entry("prefix without unique suffix", "astro_return_"),
		Entry("SQL syntax", "astro_return_a; DROP SCHEMA public"),
	)
})

var _ = Describe("InspectSQLite", func() {
	It("validates and hashes Unicode, newlines, and microsecond timestamps", func() {
		path, db := initializedSQLite()
		seedSQLite(db)

		summary, err := transfer.InspectSQLite(context.Background(), path)

		Expect(err).NotTo(HaveOccurred())
		Expect(summary.Habits.Rows).To(Equal(1))
		Expect(summary.Activities.Rows).To(Equal(1))
		Expect(summary.Groups.Rows).To(Equal(1))
		Expect(summary.Memberships.Rows).To(Equal(1))
		Expect(summary.Activities.SHA256).To(HaveLen(64))
	})

	It("includes descriptions in the field hash", func() {
		path, db := initializedSQLite()
		seedSQLite(db)
		before, err := transfer.InspectSQLite(context.Background(), path)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec("UPDATE activities SET description = ?", "a different description")
		Expect(err).NotTo(HaveOccurred())

		after, err := transfer.InspectSQLite(context.Background(), path)

		Expect(err).NotTo(HaveOccurred())
		Expect(after.Activities.Rows).To(Equal(before.Activities.Rows))
		Expect(after.Activities.SHA256).NotTo(Equal(before.Activities.SHA256))
	})

	It("refuses a database without the commissioned schema", func() {
		path := filepath.Join(GinkgoT().TempDir(), "uncommissioned.db")
		db, err := sql.Open("sqlite", path)
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Ping()).To(Succeed())
		Expect(db.Close()).To(Succeed())

		_, err = transfer.InspectSQLite(context.Background(), path)

		Expect(err).To(MatchError(transfer.ErrValidation))
	})

	It("refuses unexpected columns that the transfer would omit", func() {
		path, db := initializedSQLite()
		_, err := db.Exec("ALTER TABLE habits ADD COLUMN unexpected TEXT")
		Expect(err).NotTo(HaveOccurred())

		_, err = transfer.InspectSQLite(context.Background(), path)

		Expect(err).To(MatchError(transfer.ErrValidation))
	})

	It("refuses timestamp precision that Postgres cannot preserve", func() {
		path, db := initializedSQLite()
		seedSQLite(db)
		_, err := db.Exec("UPDATE activities SET created_at = ?", "2020-02-29 23:59:58.123456789+00:00")
		Expect(err).NotTo(HaveOccurred())

		_, err = transfer.InspectSQLite(context.Background(), path)

		Expect(err).To(MatchError(transfer.ErrInvalidData))
	})
})

func initializedSQLite() (string, *sql.DB) {
	GinkgoHelper()
	path := filepath.Join(GinkgoT().TempDir(), "astro.db")
	Expect(astrosqlite.Initialize(context.Background(), path)).To(Succeed())
	orm, err := astrosqlite.Open(astrosqlite.Config{Path: path})
	Expect(err).NotTo(HaveOccurred())
	db, err := orm.DB()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(db.Close()).To(Succeed()) })

	return path, db
}

func seedSQLite(db *sql.DB) {
	GinkgoHelper()
	_, err := db.Exec("INSERT INTO habits (id,user_id,name) VALUES (?,?,?)", habitID, userID, "déjà vu\n\"read\"")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO activities (id,habit_id,description,created_at) VALUES (?,?,?,?)", activityID, habitID, "\"quoted\"\nactivity 🪴", "2020-02-29 23:59:58.123456+00:00")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO groups (id,user_id,name) VALUES (?,?,?)", groupID, userID, "learning\n読書")
	Expect(err).NotTo(HaveOccurred())
	_, err = db.Exec("INSERT INTO groups_habits (group_id,habit_id,user_id) VALUES (?,?,?)", groupID, habitID, userID)
	Expect(err).NotTo(HaveOccurred())
}

const (
	userID     = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
	habitID    = "11767d2e-57f8-41c5-b34f-7b845a084d63"
	groupID    = "22767d2e-57f8-41c5-b34f-7b845a084d63"
	activityID = "44767d2e-57f8-41c5-b34f-7b845a084d63"
)
