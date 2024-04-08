package usecase

func (u *URLProcessor) Ping() error {
	return u.Repo.Ping()
}
