package user

import "liokoredu/application/models"

type UseCase interface {
	//GetId(token string) (uint64, error)
	StoreSession(uid uint64) (string, error)
	CheckSession(token string) (uint64, error)
	DeleteSession(token string) error
	GetUserByUsernameOrEmail(username string, email string) (*models.User, error)
	GetUserByUid(uid uint64) (*models.User, error)
	CreateUser(usr models.User) (uint64, error)
	LoginUser(usr models.UserAuth) (uint64, error)
	UpdateUser(uid uint64, usr models.UserUpdate) error
	UpdateUserAvatar(uid uint64, usr *models.Avatar) error
	UpdatePassword(uid uint64, data models.PasswordNew) error
}
