package user

import "liokoredu/application/models"

type Repository interface {
	//GetId(token string) (uint64, error)
	StoreSession(token string, uid uint64) error
	CheckSession(token string) (*uint64, error)
	DeleteSession(token string) error
	GetUserByUsernameOrEmail(username string, email string) (*models.User, error)
	GetUserByEmailSubmitted(email string) (*models.Users, error)
	GetUserByUid(uid uint64) (*models.User, error)
	CheckUser(usr models.UserAuth) (*models.User, error)
	InsertUser(usr models.User) (uint64, error)
	UpdateUser(uid uint64, usr models.UserUpdate) error
	UpdateUserAvatar(uid uint64, usr *models.Avatar) error
	UpdatePassword(uid uint64, newPassword string) error
}
