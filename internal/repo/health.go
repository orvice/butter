package repo

type HealthRepository struct{}

func NewHealthRepository() *HealthRepository {
	return &HealthRepository{}
}

func (r *HealthRepository) ServiceName() string {
	return "butter"
}
