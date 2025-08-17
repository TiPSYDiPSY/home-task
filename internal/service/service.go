package service

import "github.com/TiPSYDiPSY/home-task/internal/db"

type Container struct {
	UserService UserService
}

func NewContainer(ds *db.PostgresDBDataStore) Container {
	return Container{
		UserService: newUserService(ds),
	}
}
