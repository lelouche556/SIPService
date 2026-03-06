package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"SIP/internal/dto"
	"SIP/internal/service"
	"SIP/internal/util"
)

type UserController struct {
	userService *service.UserService
}

var (
	userControllerOnce sync.Once
	userControllerInst *UserController
)

func NewUserController(userService *service.UserService) *UserController {
	userControllerOnce.Do(func() {
		userControllerInst = &UserController{userService: userService}
	})
	return userControllerInst
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	user, err := c.userService.CreateUser(r.Context(), req.UserID, req.Name)
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	util.WriteJSON(w, http.StatusCreated, user)
}

func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := util.ParsePagination(r.URL.Query().Get("offset"), r.URL.Query().Get("limit"))
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	users, err := c.userService.ListUsers(r.Context())
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	util.WriteJSON(w, http.StatusOK, util.PaginateSlice(users, offset, limit))
}
