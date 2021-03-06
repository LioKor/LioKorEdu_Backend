package task

import "liokoredu/application/models"

type Repository interface {
	GetTask(id uint64) (*models.TaskSQL, error)
	GetPages() (int, error)
	GetTasks(page int, count int) (*models.ShortTasks, error)
	GetSolvedTasks(uid uint64, page int, count int) (*models.ShortTasks, error)
	GetUnsolvedTasks(uid uint64, page int, count int) (*models.ShortTasks, error)
	IsCleared(taskId uint64, uid uint64) (bool, error)
	GetUserTasks(uid uint64, page int, count int) (*models.ShortTasks, error)
	CreateTask(t *models.TaskSQL) (uint64, error)
	DeleteTask(id uint64, uid uint64) error
	UpdateTask(t *models.TaskSQL) error
	MarkTaskDone(id uint64, uid uint64) error
	FindTasks(str string, page int, count int) (*models.ShortTasks, error)
	FindTasksFull(str string, useSolved bool, solved bool, useMine bool, mine bool, uid uint64, page int, count int) (*models.ShortTasks, int, error)
}
