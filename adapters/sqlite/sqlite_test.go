package sqlite_test

import (
	"astro/adapters/sqlite"
	"astro/habit"
	"astro/internal/transfer"
	"context"
	"database/sql"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSQLite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SQLite storage")
}

var _ = Describe("SQLite storage", func() {
	It("opens an explicitly initialized database", func() {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		Expect(sqlite.Initialize(context.Background(), path)).To(Succeed())

		db, err := sqlite.Open(sqlite.Config{Path: path})

		Expect(err).NotTo(HaveOccurred())
		pool, err := db.DB()
		Expect(err).NotTo(HaveOccurred())
		Expect(pool.Close()).To(Succeed())
	})

	It("refuses missing storage without creating a database", func() {
		path := filepath.Join(GinkgoT().TempDir(), "missing.db")

		_, err := sqlite.Open(sqlite.Config{Path: path})

		Expect(err).To(HaveOccurred())
		_, err = os.Stat(path)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	DescribeTable("refuses uncommissioned storage", func(contents string) {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		Expect(os.WriteFile(path, []byte(contents), 0600)).To(Succeed())

		_, err := sqlite.Open(sqlite.Config{Path: path})

		Expect(err).To(HaveOccurred())
	}, Entry("empty file", ""), Entry("corrupt file", "not a SQLite database"))

	It("refuses to initialize an existing file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		original := []byte("keep this file")
		Expect(os.WriteFile(path, original, 0600)).To(Succeed())

		err := sqlite.Initialize(context.Background(), path)

		Expect(err).To(HaveOccurred())
		contents, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(contents).To(Equal(original))
	})
})

var _ = Describe("SQLite repositories", func() {
	var pool *sql.DB
	var db *gorm.DB
	var path string
	var habits *habit.HabitSQLRepository
	var groups *habit.GroupSQLRepository
	var activities *habit.ActivitySQLRepository
	const userID = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "astro.db")
		Expect(sqlite.Initialize(context.Background(), path)).To(Succeed())
		var err error
		db, err = sqlite.Open(sqlite.Config{Path: path})
		Expect(err).NotTo(HaveOccurred())
		pool, err = db.DB()
		Expect(err).NotTo(HaveOccurred())
		habits = habit.NewHabitSQLRepository(db)
		groups = habit.NewGroupSQLRepository(db)
		activities = habit.NewActivitySQLRepository(db)
	})
	AfterEach(func() { Expect(pool.Close()).To(Succeed()) })

	DescribeTable("normalizes accepted owner spellings", func(owner string) {
		created, err := habits.Create(context.Background(), habit.CreateHabitDTO{UserID: owner, Name: "read"})
		Expect(err).NotTo(HaveOccurred())

		found, err := habits.Find(context.Background(), habit.FindHabitDTO{HabitID: strings.ToUpper(created.ID), UserID: userID})

		Expect(err).NotTo(HaveOccurred())
		Expect(created.UserID).To(Equal(userID))
		Expect(found).To(Equal(created))
		_, err = transfer.InspectSQLite(context.Background(), path)
		Expect(err).NotTo(HaveOccurred())
	}, Entry("uppercase", "26B67B16-F8E7-4686-8C78-BC7F5A70ED1A"), Entry("compact", "26b67b16f8e746868c78bc7f5a70ed1a"))

	It("updates using an alternate UUID spelling", func() {
		created, err := habits.Create(context.Background(), habit.CreateHabitDTO{UserID: userID, Name: "read"})
		Expect(err).NotTo(HaveOccurred())

		err = habits.Update(context.Background(), habit.UpdateHabitDTO{HabitID: strings.ToUpper(created.ID), Name: "learn"})

		Expect(err).NotTo(HaveOccurred())
		found, err := habits.Find(context.Background(), habit.FindHabitDTO{HabitID: created.ID, UserID: userID})
		Expect(err).NotTo(HaveOccurred())
		created.Name = "learn"
		Expect(found).To(Equal(created))
	})

	DescribeTable("rejects malformed identifiers", func(operation func() error) {
		Expect(operation()).To(MatchError(habit.ErrRepository))
	},
		Entry("habit owner", func() error {
			_, err := habits.Create(context.Background(), habit.CreateHabitDTO{UserID: "invalid", Name: "read"})
			return err
		}),
		Entry("habit lookup", func() error {
			_, err := habits.Find(context.Background(), habit.FindHabitDTO{HabitID: "invalid", UserID: userID})
			return err
		}),
		Entry("habit listing owner", func() error { _, err := habits.List(context.Background(), "invalid"); return err }),
		Entry("habit deletion", func() error {
			return habits.Delete(context.Background(), habit.FindHabitDTO{HabitID: "invalid", UserID: userID})
		}),
		Entry("activity parent", func() error {
			_, err := activities.Add(context.Background(), habit.Habit{ID: "invalid"}, habit.AddActivityDTO{})
			return err
		}),
		Entry("activity lookup", func() error {
			_, err := activities.Find(context.Background(), habit.FindActivityDTO{HabitID: userID, ActivityID: "invalid"})
			return err
		}),
		Entry("activity update", func() error {
			_, err := activities.Update(context.Background(), habit.UpdateActivityDTO{ActivityID: "invalid"})
			return err
		}),
		Entry("activity deletion", func() error { return activities.Delete(context.Background(), habit.Activity{ID: "invalid"}) }),
		Entry("group owner", func() error {
			_, err := groups.Create(context.Background(), habit.CreateGroupDTO{UserID: "invalid", Name: "learning"})
			return err
		}),
		Entry("group lookup", func() error {
			_, err := groups.Find(context.Background(), habit.FindGroupDTO{GroupID: "invalid", UserID: userID})
			return err
		}),
		Entry("group deletion", func() error { return groups.Delete(context.Background(), habit.Group{ID: "invalid", UserID: userID}) }),
	)

	It("retains durability and foreign key settings after replacing a connection", func() {
		pool.SetMaxIdleConns(0)
		var mode string
		var synchronous, foreignKeys, timeout int

		Expect(pool.QueryRow("PRAGMA journal_mode").Scan(&mode)).To(Succeed())
		Expect(pool.QueryRow("PRAGMA synchronous").Scan(&synchronous)).To(Succeed())
		Expect(pool.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)).To(Succeed())
		Expect(pool.QueryRow("PRAGMA busy_timeout").Scan(&timeout)).To(Succeed())

		Expect(mode).To(Equal("wal"))
		Expect([]int{synchronous, foreignKeys, timeout}).To(Equal([]int{2, 1, 1000}))
	})

	It("rejects memberships between different users", func() {
		ctx := context.Background()
		h, err := habits.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"})
		Expect(err).NotTo(HaveOccurred())
		g, err := groups.Create(ctx, habit.CreateGroupDTO{UserID: "f5157ef6-1277-4e26-a830-2e146dd980dc", Name: "learning"})
		Expect(err).NotTo(HaveOccurred())

		_, err = pool.Exec("INSERT INTO groups_habits(group_id,habit_id,user_id) VALUES(?,?,?)", g.ID, h.ID, userID)

		Expect(err).To(HaveOccurred())
	})

	It("cascades a habit deletion to its activities and memberships", func() {
		ctx := context.Background()
		h, err := habits.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"})
		Expect(err).NotTo(HaveOccurred())
		g, err := groups.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "learning"})
		Expect(err).NotTo(HaveOccurred())
		_, err = activities.Add(ctx, h, habit.AddActivityDTO{Desc: "chapter", Time: time.Now().UTC()})
		Expect(err).NotTo(HaveOccurred())
		Expect(groups.Join(ctx, h, g)).To(Succeed())

		Expect(habits.Delete(ctx, habit.FindHabitDTO{HabitID: h.ID, UserID: userID})).To(Succeed())

		var activityCount, membershipCount int
		Expect(pool.QueryRow("SELECT count(*) FROM activities").Scan(&activityCount)).To(Succeed())
		Expect(pool.QueryRow("SELECT count(*) FROM groups_habits").Scan(&membershipCount)).To(Succeed())
		Expect([]int{activityCount, membershipCount}).To(Equal([]int{0, 0}))
	})

	It("preserves imported identifiers and timestamp precision after reopening", func() {
		h := habit.Habit{ID: "98f21686-c9ca-4555-bd22-21b138eb6152", UserID: userID, Name: "déjà vu"}
		a := habit.Activity{ID: "da01d04c-867a-4e6f-949a-ad9dab0adbf3", HabitID: h.ID, Desc: "\"quoted\"\nline", CreatedAt: time.Date(2020, 2, 29, 23, 59, 58, 123456000, time.UTC)}
		Expect(db.Create(&h).Error).To(Succeed())
		Expect(db.Create(&a).Error).To(Succeed())
		pool.SetMaxIdleConns(0)

		found, err := habits.Find(context.Background(), habit.FindHabitDTO{HabitID: h.ID, UserID: userID})

		Expect(err).NotTo(HaveOccurred())
		h.Activities = []habit.Activity{a}
		Expect(found).To(Equal(h))
	})

	It("persists concurrent check-ins to the same habit exactly once", func() {
		ctx := context.Background()
		h, err := habits.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"})
		Expect(err).NotTo(HaveOccurred())
		failures := make(chan error, 200)
		var workers sync.WaitGroup

		for range 8 {
			workers.Go(func() {
				for range 25 {
					_, err := activities.Add(ctx, h, habit.AddActivityDTO{Desc: "chapter", Time: time.Now().UTC()})
					if err != nil {
						failures <- err
					}
				}
			})
		}
		workers.Wait()
		close(failures)

		Expect(failures).To(BeClosed())
		found, err := habits.Find(ctx, habit.FindHabitDTO{HabitID: h.ID, UserID: userID})
		Expect(err).NotTo(HaveOccurred())
		Expect(found.Activities).To(HaveLen(200))
	})

	It("returns a bounded repository error when another writer holds the file", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		h, err := habits.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"})
		Expect(err).NotTo(HaveOccurred())
		external, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(1000)")
		Expect(err).NotTo(HaveOccurred())
		defer func() { Expect(external.Close()).To(Succeed()) }()
		external.SetMaxOpenConns(1)
		_, err = external.ExecContext(ctx, "BEGIN IMMEDIATE")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _, err := external.Exec("ROLLBACK"); Expect(err).NotTo(HaveOccurred()) }()

		_, err = activities.Add(ctx, h, habit.AddActivityDTO{Desc: "blocked", Time: time.Now().UTC()})

		Expect(err).To(MatchError(habit.ErrRepository))
	})

	It("rejects a commissioned file with a missing application table", func() {
		_, err := pool.Exec("DROP TABLE activities")
		Expect(err).NotTo(HaveOccurred())

		err = sqlite.CheckFile(context.Background(), path)

		Expect(err).To(MatchError(sqlite.ErrUncommissioned))
	})

	DescribeTable("rejects incomplete commissioned constraints", func(statement string) {
		_, err := pool.Exec(statement)
		Expect(err).NotTo(HaveOccurred())

		err = sqlite.CheckFile(context.Background(), path)

		Expect(err).To(MatchError(sqlite.ErrUncommissioned))
		_, err = sqlite.Open(sqlite.Config{Path: path})
		Expect(err).To(MatchError(sqlite.ErrUncommissioned))
	},
		Entry("tenant foreign keys", "DROP TABLE groups_habits; CREATE TABLE groups_habits(group_id TEXT NOT NULL, habit_id TEXT NOT NULL, user_id TEXT NOT NULL, PRIMARY KEY(group_id,habit_id))"),
		Entry("group tenant foreign key", "DROP TABLE groups_habits; CREATE TABLE groups_habits(group_id TEXT NOT NULL, habit_id TEXT NOT NULL, user_id TEXT NOT NULL, PRIMARY KEY(group_id,habit_id), FOREIGN KEY(habit_id,user_id) REFERENCES habits(id,user_id) ON DELETE CASCADE ON UPDATE CASCADE)"),
		Entry("habit tenant foreign key", "DROP TABLE groups_habits; CREATE TABLE groups_habits(group_id TEXT NOT NULL, habit_id TEXT NOT NULL, user_id TEXT NOT NULL, PRIMARY KEY(group_id,habit_id), FOREIGN KEY(group_id,user_id) REFERENCES groups(id,user_id) ON DELETE CASCADE ON UPDATE CASCADE)"),
		Entry("cascade", "DROP TABLE activities; CREATE TABLE activities(id TEXT NOT NULL PRIMARY KEY, habit_id TEXT NOT NULL REFERENCES habits(id), description TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP); CREATE INDEX idx_activity_habit ON activities(habit_id)"),
		Entry("activity primary key", "DROP TABLE activities; CREATE TABLE activities(id TEXT NOT NULL, habit_id TEXT NOT NULL REFERENCES habits(id) ON DELETE CASCADE, description TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP); CREATE INDEX idx_activity_habit ON activities(habit_id)"),
		Entry("activity index", "DROP INDEX idx_activity_habit"),
	)

	It("rejects orphaned rows when checking a restored file", func() {
		external, err := sql.Open("sqlite", "file:"+path)
		Expect(err).NotTo(HaveOccurred())
		defer func() { Expect(external.Close()).To(Succeed()) }()
		_, err = external.Exec("INSERT INTO activities(id,habit_id,description,created_at) VALUES('orphan','missing','invalid','2026-10-18 12:00:00')")
		Expect(err).NotTo(HaveOccurred())

		err = sqlite.CheckFile(context.Background(), path)

		Expect(err).To(MatchError(sqlite.ErrUncommissioned))
	})
})

var _ = Describe("SQLite initialization", func() {
	DescribeTable("refuses orphaned sidecars", func(suffix string) {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		Expect(os.WriteFile(path+suffix, []byte("retained"), 0600)).To(Succeed())

		err := sqlite.Initialize(context.Background(), path)

		Expect(err).To(HaveOccurred())
		data, err := os.ReadFile(path + suffix)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("retained"))
	}, Entry("WAL", "-wal"), Entry("shared memory", "-shm"))
})
