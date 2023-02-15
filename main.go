package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"taskManager/controllers"
	"taskManager/entity"
)

func main() {

	ctx := context.Background()
	db, err := firestore.NewClient(ctx, entity.GCP_PROJECT)
	if err != nil {
		log.Fatal(ctx, err.Error())
	}

	con := controllers.NewController(db)

	r := gin.Default()

	r.GET("list", con.ListAll)
	r.POST("create", con.CreateTask)
	r.DELETE("delete", con.DeleteTask)

	log.Fatal(r.Run(":9000"))
}
