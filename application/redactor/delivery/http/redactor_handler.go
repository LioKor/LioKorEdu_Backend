package http

import (
	"liokoredu/application/models"
	"liokoredu/application/server/middleware"
	"liokoredu/pkg/constants"
	"liokoredu/pkg/generators"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/mailru/easyjson"
)

type RedactorHandler struct {
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// это стор, где хранятся сессии
var subscriptions = make(map[string]*Session)

func CreateRedactorHandler(e *echo.Echo, a middleware.Auth) {

	redactorHandler := RedactorHandler{}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	e.POST("/api/v1/redactor", redactorHandler.CreateConnection)
	e.GET("/api/v1/ws/redactor/:id", redactorHandler.ConnectToRoom)

}

func (rh *RedactorHandler) CreateConnection(c echo.Context) error {
	defer c.Request().Body.Close()

	sln := &models.SolutionFile{}

	if err := easyjson.UnmarshalFromReader(c.Request().Body, sln); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusTeapot, err.Error())
	}
	log.Println(sln)

	roomId, _ := createRoom(sln.SourceCode)

	return c.JSON(http.StatusOK, &models.IdValue{Id: roomId})
}

func (rh *RedactorHandler) ConnectToRoom(c echo.Context) error {
	defer c.Request().Body.Close()

	id := c.Param(constants.IdKey)
	log.Println(id)
	s := getRoom(id)
	if s == nil {
		return c.JSON(http.StatusNotFound, nil)
	}
	serveWs(c, s)

	return nil
}

func createRoom(code string) (string, *Session) {
	roomId := generators.RandStringRunes(constants.WSLength)
	session := NewSession(code)
	go session.HandleEvents()
	go session.PingPong()
	go session.CheckLive(roomId, subscriptions)
	subscriptions[roomId] = session
	return roomId, session
}

func getRoom(id string) *Session {
	return subscriptions[id]
}

func serveWs(c echo.Context, s *Session) {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("error serving ws:", err)
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	NewConnection(s, conn).Handle()
}
