package controllers

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"taskManager/entity"
	"taskManager/repository"
	"taskManager/server"
)

type Controller struct {
	db *firestore.Client
}

func NewController(db *firestore.Client) *Controller {
	return &Controller{db: db}
}

func (c *Controller) ListAll(ctx *gin.Context) {
	servers, err := repository.ListServers(ctx, c.db)
	tasks, err := repository.ListTasks(ctx, c.db)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to GET data from Firestore",
			})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusOK,
		gin.H{
			"server info   ": servers,
			"task info   ":   tasks,
		})
	return
}

func (c *Controller) CreateTask(ctx *gin.Context) {

	// It only takes the size of the task this user wants to run
	var body *entity.Task
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error message": "failed to parse JSON data",
			})
		return
	}

	servers, err := repository.ListServers(ctx, c.db)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to GET data from Firestore",
			})
		return
	}
	limit, softLimit, err := repository.GetLimit(ctx, c.db)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to GET data from Firestore",
			})
		return
	}

	// if there is still space to add task within the soft max.
	// Look all the servers to check if there is any space
	for _, s := range servers {
		totalSize := s.Size + body.Size
		if totalSize < softLimit {
			user := entity.Task{ID: uuid.New().String(), Size: body.Size, ServerID: s.ID}
			err = repository.AddTask(ctx, c.db, s, user)
			if err != nil {
				fmt.Println(err.Error())
				ctx.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error message": "failed to POST data to Firestore",
					})
				return
			} else {
				ctx.AbortWithStatusJSON(http.StatusOK,
					gin.H{
						"message": "Task was added",
					})
				return
			}
		}
	}

	// If all the servers reached soft max, it looks for the server that can
	// still fit the task size. Then it creates new server for next user.
	for _, s := range servers {
		totalSize := s.Size + body.Size
		// if there is still space to add task but over the soft limit, it creates new server
		if totalSize >= softLimit && totalSize <= limit {
			user := entity.Task{ID: uuid.New().String(), Size: body.Size, ServerID: s.ID}
			err = repository.AddTask(ctx, c.db, s, user)
			if err != nil {
				fmt.Println(err.Error())
				ctx.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error message": "failed to POST data to Firestore",
					})
				return
			}

			err = repository.AddServer(ctx, c.db, s.ID+1)
			if err != nil {
				fmt.Println(err.Error())
				ctx.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error message": "failed to POST data to Firestore",
					})
				return
			}

			err = server.CreateServer(ctx, "server"+strconv.Itoa(s.ID+1))
			if err != nil {
				fmt.Println(err.Error())
				ctx.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error message": "failed to POST data to Google Cloud Engine",
					})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusOK,
				gin.H{
					"message": "Task was added",
				})
			return
		}

	}

	ctx.AbortWithStatusJSON(http.StatusInternalServerError,
		gin.H{
			"error message": "server was not found",
		})
	return
}

func (c *Controller) DeleteTask(ctx *gin.Context) {

	// It only takes the ID of the user to deletes the task as the user wants
	var body *entity.Task
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error message": "failed to parse JSON data",
			})
		return
	}

	// Delete the specific task
	err := repository.DeleteTask(ctx, c.db, body.ID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to DELETE data to Firestore",
			})
		return
	}

	// If there is a server that doesn't run the task and one server that its size is
	// lower than soft limit, delete the empty server.
	_, softLimit, err := repository.GetLimit(ctx, c.db)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to GET data from Firestore",
			})
		return
	}

	servers, err := repository.ListServers(ctx, c.db)
	if err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{
				"error message": "failed to GET data from Firestore",
			})
		return
	}
	emptyServerName := ""
	hasLowSizeServer := false
	for _, s := range servers {
		if s.Size == 0 {
			emptyServerName = "server" + strconv.Itoa(s.ID)
		} else if s.Size < softLimit {
			hasLowSizeServer = true
		}
	}
	if emptyServerName != "" && hasLowSizeServer {
		err = server.DeleteServer(ctx, emptyServerName)
		if err != nil {
			fmt.Println(err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{
					"error message": "failed to DELETE data to GCE",
				})
			return
		}
	}

	ctx.AbortWithStatusJSON(http.StatusOK,
		gin.H{
			"message": "Task was deleted",
		})
	return

}
