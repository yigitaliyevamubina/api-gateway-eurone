package v1

import (
	"encoding/json"
	er "fourth-exam/api_gateway_evrone/api/errors"
	"fourth-exam/api_gateway_evrone/api/handlers"
	"fourth-exam/api_gateway_evrone/api/middleware"
	"fourth-exam/api_gateway_evrone/api/models"
	pb "fourth-exam/api_gateway_evrone/genproto/user_service"
	grpserviceclient "fourth-exam/api_gateway_evrone/internal/infrastructure/grp_service_client"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"
	"fourth-exam/api_gateway_evrone/internal/usecase/event"

	"net/http"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userHandler struct {
	handlers.BaseHandler
	logger   *zap.Logger
	config   *config.Config
	service  grpserviceclient.ServiceClient
	enforcer *casbin.CachedEnforcer
	producer event.BrokerProducer
}

func NewUserHandler(option *handlers.HandlerOption) http.Handler {
	handler := userHandler{
		logger:   option.Logger,
		config:   option.Config,
		service:  option.Service,
		enforcer: option.Enforcer,
		producer: option.BrokerProducer,
	}

	handler.Cache = option.Cache
	handler.Client = option.Service
	handler.config = option.Config

	policies := [][]string{
		{"unauthorized", "/v1/user/create", "GET"},
		{"unauthorized", "/v1/user/update/{id}", "PUT"},
		{"unauthorized", "/v1/user/get/{id}", "GET"},
		{"unauthorized", "/v1/user/delete/{id}", "DELETE"},
		{"unauthorized", "/v1/users", "GET"},
	}
	for _, policy := range policies {
		_, err := option.Enforcer.AddPolicy(policy)
		if err != nil {
			option.Logger.Error("error during app enforcer add policies", zap.Error(err))
		}
	}

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		// auth
		r.Use(middleware.Authorizer(option.Enforcer, option.Logger))

		// user
		r.Post("/create", handler.Create())
		r.Put("/update/{id}", handler.Update())
		r.Get("/get/{id}", handler.Get())
		r.Delete("/delete/{id}", handler.Delete())
		r.Get("/users", handler.List())
	})
	return router
}

// User create
// @Security ApiKeyAuth
// @Router /v1/user/create [POST]
// @Summary Create user
// @Description Create user
// @Tags User
// @Accept json
// @Produce json
// @Param user-info body models.User true "user info"
// @Success 200 {object} models.User
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *userHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}
		user.Id = uuid.New().String()

		userProto := &pb.User{
			Id:           user.Id,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.Email,
			Password:     user.Password,
			Bio:          user.Bio,
			Website:      user.Website,
			IsActive:     user.IsActive,
			RefreshToken: user.RefreshToken,
		}

		err = h.producer.ProduceUserInfoToKafka(ctx, "api.user.created", userProto)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}

		// _, err = h.service.UserService().Create(ctx, userProto)
		// if err != nil {
		// 	render.Render(w, r, er.Error(err))
		// 	return
		// }

		render.JSON(w, r, user)
	}
}

// User update
// @Security ApiKeyAuth
// @Router /v1/user/update/{id} [PUT]
// @Summary Update user
// @Description Update user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "user id"
// @Param user-info body models.User true "user info"
// @Success 200 {object} models.User
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *userHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}
		user.Id = chi.URLParam(r, "id")

		_, err = h.service.UserService().Update(ctx, &pb.User{
			Id:           user.Id,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.Email,
			Password:     user.Password,
			Bio:          user.Bio,
			Website:      user.Website,
			IsActive:     user.IsActive,
			RefreshToken: user.RefreshToken,
		})
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}

		render.JSON(w, r, user)
	}
}

// GetUser
// @Security ApiKeyAuth
// @Router /v1/user/get/{id} [GET]
// @Summary Get user
// @Description Get user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} models.User
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *userHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId := chi.URLParam(r, "id")

		user, err := h.service.UserService().Get(ctx, &pb.GetRequest{UserId: userId})
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}

		userModel := models.User{
			Id:           user.Id,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.Email,
			Password:     user.Password,
			Bio:          user.Bio,
			Website:      user.Website,
			IsActive:     user.IsActive,
			RefreshToken: user.RefreshToken,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}

		render.JSON(w, r, userModel)
	}
}

// DeleteUser
// @Security ApiKeyAuth
// @Router /v1/user/delete/{id} [DELETE]
// @Summary Delete user
// @Description Delete user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} models.Message
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *userHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId := chi.URLParam(r, "id")

		_, err := h.service.UserService().Delete(ctx, &pb.GetRequest{UserId: userId})
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}

		render.JSON(w, r, models.Message{
			Message: "user was successfully deleted",
		})
	}
}

// User List
// @Security ApiKeyAuth
// @Router /v1/user/users [GET]
// @Summary List users
// @Description List users
// @Tags User
// @Accept json
// @Produce json
// @Param request query models.GetListFilter true "request"
// @Success 200 {object} models.Users
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *userHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		page := r.URL.Query().Get("page")
		pageToInt, err := strconv.Atoi(page)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}
		limit := r.URL.Query().Get("limit")
		limitToInt, err := strconv.Atoi(limit)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}
		orderBy := r.URL.Query().Get("order_by")

		var (
			req pb.GetListFilter
		)
		if orderBy != "" {
			req.OrderBy = orderBy
		}
		if pageToInt != 0 {
			req.Page = int64(pageToInt)
		} else {
			req.Page = 1
		}
		if limitToInt != 0 {
			req.Limit = int64(limitToInt)
		} else {
			req.Limit = 10
		}

		users, err := h.service.UserService().List(ctx, &req)
		if err != nil {
			render.Render(w, r, er.Error(err))
			return
		}

		var respUsers models.Users
		for _, user := range users.Users {
			respUsers.Users = append(respUsers.Users, &models.User{
				Id:           user.Id,
				Username:     user.Username,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
				Email:        user.Email,
				Password:     user.Password,
				Bio:          user.Bio,
				Website:      user.Website,
				IsActive:     user.IsActive,
				RefreshToken: user.RefreshToken,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
			})
		}
		respUsers.Count = users.Count

		render.JSON(w, r, respUsers)
	}
}
