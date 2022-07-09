package habits_test

import (
	"astro/config"
	"astro/habit"
	"astro/http/fiber"
	"astro/http/habits"
	"astro/logger"
	"astro/postgres"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/habits suite")
}

var _ = Describe("/habits", func() {
	var app *fxtest.App

	BeforeEach(func() {
		var svc *habit.SQLHabitRepository

		app = fxtest.New(
			GinkgoT(),
			fx.NopLogger,
			config.Module,
			logger.Module,
			fiber.Module,
			postgres.Module,
			habit.Module,
			habits.Providers,
			fx.Populate(&svc),
		)
		app.RequireStart()
		svc.DeleteAll(context.Background())
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("GET", func() {
		It("returns a list of habits", func() {
			Must(NewDriver().Create("read"))

			data := Must2(NewDriver().List())

			Expect(data).To(HaveLen(1))
			Expect(data[0].Name).To(Equal("read"))
			Expect(data[0].Activities).To(HaveLen(0))
		})
	})

	Describe("POST", func() {
		It("returns status created", func() {
			res, _ := NewAPI().Create("read")
			Expect(res.StatusCode).To(Equal(http.StatusCreated))
		})

		Describe("without name", func() {
			It("return bad request", func() {
				res, _ := NewAPI().Create("")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

type API struct{}

func NewAPI() *API {
	return &API{}
}

func (a API) List() (*http.Response, error) {
	return http.Get("http://localhost:3000/habits")
}

func (a API) Create(name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits?name=%s", name)
	return http.Post(url, "application/text", strings.NewReader(""))
}

type Driver struct {
	api *API
}

func NewDriver() *Driver {
	return &Driver{NewAPI()}
}

func (d *Driver) List() ([]habit.Habit, error) {
	data := []habit.Habit{}

	res, err := d.api.List()
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	str, err := io.ReadAll(res.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(str, &data)

	return data, err
}

func (d *Driver) Create(name string) error {
	_, err := d.api.Create(name)
	return err
}
