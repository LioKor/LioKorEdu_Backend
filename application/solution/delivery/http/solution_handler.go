package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"liokoredu/application/models"
	"liokoredu/application/solution"
	"liokoredu/application/task"
	"liokoredu/application/user"
	"liokoredu/pkg/constants"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/mailru/easyjson"
)

type SolutionHandler struct {
	UseCase  solution.UseCase
	TUseCase task.UseCase
	uuc      user.UseCase
}

func CreateSolutionHandler(e *echo.Echo,
	uc solution.UseCase, tuc task.UseCase, uuc user.UseCase) {
	solutionHandler := SolutionHandler{
		UseCase:  uc,
		TUseCase: tuc,
		uuc:      uuc,
	}
	e.POST("/api/v1/tasks/:id/solutions", solutionHandler.PostSolution)
	e.POST("/api/v1/solutions/update/:id", solutionHandler.UpdateSolution)
	e.GET("/api/v1/tasks/:id/solutions", solutionHandler.GetSolutions)
	e.GET("/api/v1/tasks/:taskId/solutions/:solutionId", solutionHandler.getSolution)
	e.PUT("/api/v1/tasks/:taskId/solutions/:solutionId", solutionHandler.rerunSolution)
	e.DELETE("/api/v1/tasks/:taskId/solutions/:solutionId", solutionHandler.deleteSolution)
}

func (sh SolutionHandler) PostSolution(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	cookie, err := c.Cookie(constants.SessionCookieName)
	if err != nil && cookie != nil {
		log.Println("user handler: PostSolution: error getting cookie")
		return echo.NewHTTPError(http.StatusBadRequest, "error getting cookie")
	}

	if cookie == nil {
		log.Println("user handler: PostSolution: no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	uid, err := sh.uuc.CheckSession(cookie.Value)
	if err != nil {
		return err
	}

	if uid == 0 {
		log.Println("user handler: PostSolution: uid 0")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	sln := &models.Solution{}
	id := c.Param(constants.IdKey)
	iid, _ := strconv.ParseUint(string(id), 10, 64)

	if err := easyjson.UnmarshalFromReader(c.Request().Body, sln); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusTeapot, err.Error())
	}

	task, err := sh.TUseCase.GetTask(iid, uid, true)
	if err != nil {
		return err
	}
	testAmount := task.TestsAmount

	solId, err := sh.UseCase.InsertSolution(iid, uid, sln.SourceCode, testAmount)
	if err != nil {
		return err
	}

	ss := models.SolutionSend{
		Id:         solId,
		SourceCode: sln.SourceCode,
		Tests:      models.InputTests(task.Tests),
	}

	reqBody, err := json.Marshal(ss)
	if err != nil {
		log.Println("user handler: postSolution: error marshaling SolutionSend", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := http.Post(constants.PythonAddress,
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}

	update := &models.SolutionUpdate{}

	_ = json.Unmarshal(body, update)
	err = sh.UseCase.UpdateSolution(solId, *update)
	if err != nil {
		return err
	}

	if update.Code == 0 {
		err = sh.TUseCase.MarkTaskDone(iid, uid)
		if err != nil {
			return err
		}
	}

	ans := &models.ReturnId{Id: solId}
	if _, err = easyjson.MarshalToWriter(ans, c.Response().Writer); err != nil {
		log.Println(c, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func (sh SolutionHandler) UpdateSolution(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	id := c.Param(constants.IdKey)
	uid, _ := strconv.ParseUint(string(id), 10, 64)

	info := &models.SolutionUpdate{}
	if err := easyjson.UnmarshalFromReader(c.Request().Body, info); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusTeapot, err.Error())
	}

	err := sh.UseCase.UpdateSolution(uid, *info)
	if err != nil {
		return err
	}

	return nil
}

func (sh SolutionHandler) rerunSolution(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	cookie, err := c.Cookie(constants.SessionCookieName)
	if err != nil && cookie != nil {
		log.Println("user handler: rerunSolution: error getting cookie")
		return echo.NewHTTPError(http.StatusBadRequest, "error getting cookie")
	}

	if cookie == nil {
		log.Println("user handler: rerunSolution: no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	uid, err := sh.uuc.CheckSession(cookie.Value)
	if err != nil {
		return err
	}

	if uid == 0 {
		log.Println("user handler: rerunSolution: uid 0")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	tid := c.Param(constants.TaskId)
	utid, _ := strconv.ParseUint(string(tid), 10, 64)

	solId := c.Param(constants.SolutionId)
	usolId, _ := strconv.ParseUint(string(solId), 10, 64)

	sln, err := sh.UseCase.GetSolution(usolId, utid, uid)
	if err != nil {
		return err
	}
	tsk, err := sh.TUseCase.GetTask(utid, uid, true)
	if err != nil {
		return err
	}

	ss := models.SolutionSend{
		Id:         usolId,
		SourceCode: sln.SourceCode,
		Tests:      models.InputTests(tsk.Tests),
	}

	reqBody, err := json.Marshal(ss)
	if err != nil {
		log.Println("user handler: rerunSolution: error marshaling SolutionSend", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := http.Post(constants.PythonAddress,
		"application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}

	update := &models.SolutionUpdate{}

	_ = json.Unmarshal(body, update)
	log.Println(string(body[:]))
	err = sh.UseCase.UpdateSolution(usolId, *update)
	if err != nil {
		return err
	}

	return nil
}

func (sh SolutionHandler) GetSolutions(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	cookie, err := c.Cookie(constants.SessionCookieName)
	if err != nil && cookie != nil {
		log.Println("user handler: GetSolutions: error getting cookie")
		return echo.NewHTTPError(http.StatusBadRequest, "error getting cookie")
	}

	if cookie == nil {
		log.Println("user handler: GetSolutions: no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	uid, err := sh.uuc.CheckSession(cookie.Value)
	if err != nil {
		return err
	}

	if uid == 0 {
		log.Println("user handler: GetSolutions: uid 0")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	id := c.Param(constants.IdKey)
	iid, _ := strconv.ParseUint(string(id), 10, 64)

	slns, err := sh.UseCase.GetSolutions(iid, uid)
	if err != nil {
		return err
	}

	if _, err = easyjson.MarshalToWriter(slns, c.Response().Writer); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}

func (sh SolutionHandler) getSolution(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	cookie, err := c.Cookie(constants.SessionCookieName)
	if err != nil && cookie != nil {
		log.Println("user handler: getSolution: error getting cookie")
		return echo.NewHTTPError(http.StatusBadRequest, "error getting cookie")
	}

	if cookie == nil {
		log.Println("user handler: getSolution: no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	uid, err := sh.uuc.CheckSession(cookie.Value)
	if err != nil {
		return err
	}

	if uid == 0 {
		log.Println("user handler: getSolution: uid 0")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	taskId, _ := strconv.ParseUint(string(c.Param(constants.TaskId)), 10, 64)
	solId, _ := strconv.ParseUint(string(c.Param(constants.SolutionId)), 10, 64)

	sln, err := sh.UseCase.GetSolution(solId, taskId, uid)
	if err != nil {
		return err
	}

	if _, err = easyjson.MarshalToWriter(sln, c.Response().Writer); err != nil {
		log.Println(c, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func (sh SolutionHandler) deleteSolution(c echo.Context) error {
	defer c.Request().Body.Close()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	cookie, err := c.Cookie(constants.SessionCookieName)
	if err != nil && cookie != nil {
		log.Println("user handler: deleteSolution: error getting cookie")
		return echo.NewHTTPError(http.StatusBadRequest, "error getting cookie")
	}

	if cookie == nil {
		log.Println("user handler: deleteSolution: no cookie")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	uid, err := sh.uuc.CheckSession(cookie.Value)
	if err != nil {
		return err
	}

	if uid == 0 {
		log.Println("user handler: deleteSolution: uid 0")
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	id := c.Param(constants.SolutionId)
	iid, _ := strconv.ParseUint(string(id), 10, 64)

	err = sh.UseCase.DeleteSolution(iid, uid)
	if err != nil {
		return err
	}

	return nil
}
