package savings

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// skipped: full implementation, add when router wiring is complete.
