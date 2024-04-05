package usecase

func (s *URLProcessor) Ping() error {
	return s.Repo.Ping()
}
