package user

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(u User) error {
	return s.repo.Create(u)
}

func (s *Service) ListUsers() ([]User, error) {
	return s.repo.List()
}

func (s *Service) CountUsers() (int, error) {
	return s.repo.Count()
}

func (s *Service) GetUserByEmail(email string) (*User, error) {
	return s.repo.GetByEmail(email)
}

func (s *Service) EmailExists(email string) (bool, error) {
	return s.repo.EmailExists(email)
}
