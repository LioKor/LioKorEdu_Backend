package repository

import (
	"context"
	"database/sql"
	"errors"
	"liokoredu/application/models"
	"liokoredu/application/user"
	"liokoredu/pkg/constants"
	"log"
	"strconv"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserDatabase struct {
	poolRedis *redis.Pool
	pool      *pgxpool.Pool
}

// CheckUser implements user.Repository
func (ud *UserDatabase) CheckUser(usr models.UserAuth) (*models.User, error) {
	var gotUser models.User
	err := ud.pool.QueryRow(context.Background(), `SELECT id, username, fullname, password, email
	 FROM users WHERE username = $1`, usr.Username).Scan(&gotUser.Id, &gotUser.Username, &gotUser.Fullname, &gotUser.Password,
		&gotUser.Email)

	if errors.As(err, &sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Println("user repository: CheckUser: error getting user:", err)
		return nil, err
	}

	if gotUser.Id == 0 {
		return nil, nil
	}

	return &gotUser, nil
}

// InsertUser implements user.Repository
func (ud *UserDatabase) InsertUser(usr models.User) (uint64, error) {
	var id uint64
	err := ud.pool.QueryRow(context.Background(),
		`INSERT INTO users (username, email, password, fullname, joined_date) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		usr.Username, usr.Email, usr.Password, usr.Fullname, usr.JoinedDate).Scan(&id)
	if err != nil {
		log.Println("user repository: InsertUser: error inserting user:", err)
		return 0, err
	}

	return id, nil
}

// GetUser implements user.Repository
func (ud *UserDatabase) GetUserByUsernameOrEmail(username string, email string) (*models.User, error) {
	var usrs models.Users
	err := pgxscan.Select(context.Background(), ud.pool, &usrs,
		`SELECT * FROM users WHERE username = $1 or email = $2`,
		username, email)
	if err != nil {
		log.Println("user repository: GetUserByUsernameOrEmail: error getting user:", err)
		return nil, err
	}

	if len(usrs) == 0 {
		return nil, nil
	}

	return &usrs[0], nil
}

func (ud *UserDatabase) GetUserByUid(uid uint64) (*models.User, error) {
	var usrs models.Users
	err := pgxscan.Select(context.Background(), ud.pool, &usrs,
		`SELECT * FROM users WHERE id = $1`, uid)
	if err != nil {
		log.Println("user repository: GetUserByUid: error getting user", err)
		return nil, err
	}

	if len(usrs) == 0 {
		return nil, nil
	}

	return &usrs[0], nil
}

func (ud *UserDatabase) StoreSession(token string, uid uint64) error {
	client := ud.poolRedis.Get()
	defer client.Close()

	_, err := client.Do("SET", token, strconv.FormatUint(uid, 10), "EX", constants.WeekSec)
	if err != nil {
		log.Println("user repo: storeSession: error storing session")
		return err
	}

	return nil
}

func (ud *UserDatabase) CheckSession(token string) (*uint64, error) {
	client := ud.poolRedis.Get()
	defer client.Close()

	value, err := redis.Uint64(client.Do("GET", token))
	if err != nil && value == 0 {
		return nil, nil
	}
	if err != nil {
		log.Println("user repo: CheckSession: error checking session", err)
		return nil, err
	}

	return &value, nil
}

/*
func (ud *UserDatabase) GetId(token string) (uint64, error) {
	client := ud.poolRedis.Get()
	defer client.Close()

	uid, err := client.Do("GET", token)
	if err != nil {
		log.Println("user repo: getId: error getting id")
		return 0, err
	}

	_, err = client.Do("EXPIRE", token, constants.WeekSec)
	if err != nil {
		log.Println("user repo: getId: error updating ttl")
		return uid.(uint64), err
	}

	return uid.(uint64), nil
}
*/

// DeleteSession implements user.Repository
func (ud *UserDatabase) DeleteSession(token string) error {
	client := ud.poolRedis.Get()
	defer client.Close()

	_, err := client.Do("DEL", token)
	if err != nil {
		log.Println("user repo: DeleteSession: error deleting session")
		return err
	}

	return nil
}

func NewUserDatabase(pool *redis.Pool, poolDB *pgxpool.Pool) user.Repository {
	return &UserDatabase{poolRedis: pool, pool: poolDB}
}
