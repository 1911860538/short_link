package service

type mockLifespan struct{}

func (t *mockLifespan) Startup() error {
	return nil
}

func (t *mockLifespan) Shutdown() error {
	return nil
}
