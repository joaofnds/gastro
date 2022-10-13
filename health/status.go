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

type Check struct {
	DB Status `json:"db"`
}

func (hc Check) AllUp() bool {
	return hc.DB.IsUp()
}
