package health

const (
	StatusUp   = "up"
	StatusDown = "down"
)

type Status struct {
	Status string `json:"status"`
}

func (s Status) IsUp() bool {
	return s.Status == StatusUp
}

type HealthCheck struct {
	DB Status `json:"db"`
}

func (hc HealthCheck) AllUp() bool {
	return hc.DB.IsUp()
}
