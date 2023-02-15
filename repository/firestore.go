package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/iterator"
	"taskManager/entity"
)

func ListServers(ctx context.Context, db *firestore.Client) ([]entity.Server, error) {
	var servers []entity.Server
	err := db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {
		q := db.Collection("servers").OrderByPath(firestore.FieldPath{"Size"}, firestore.Desc)
		iter := tx.Documents(q)
		defer iter.Stop()
		for {
			var s entity.Server
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			err = doc.DataTo(&s)
			if err != nil {
				return err
			}
			servers = append(servers, s)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func GetLimit(ctx context.Context, db *firestore.Client) (capacity int, border int, err error) {
	err = db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {
		df := db.Collection("threshold").Doc("threshold")
		snap, err := tx.Get(df)
		if err != nil {
			return err
		}
		capacity = int(snap.Data()["max"].(int64))
		border = int(snap.Data()["soft max"].(int64))
		return nil
	})
	if err != nil {
		return -1, -1, err
	}
	return
}

func ListTasks(ctx context.Context, db *firestore.Client) ([]entity.Task, error) {
	var tasks []entity.Task
	err := db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {
		q := db.Collection("tasks").OrderByPath(firestore.FieldPath{"ID"}, firestore.Asc)
		iter := tx.Documents(q)
		defer iter.Stop()
		for {
			var u entity.Task
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			err = doc.DataTo(&u)
			if err != nil {
				return err
			}
			tasks = append(tasks, u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func AddTask(ctx context.Context, db *firestore.Client, server entity.Server, task entity.Task) error {
	err := db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {

		// first get the Document ID for the specific server
		q := db.Collection("servers").Where("ID", "==", server.ID)
		data, err := tx.Documents(q).GetAll()
		if err != nil {
			return err
		}
		docID := data[0].Ref.ID

		// then update the specific document
		dr := db.Collection("servers").Doc(docID)
		err = tx.Set(dr, map[string]interface{}{
			"Size": server.Size + task.Size,
		}, firestore.MergeAll)
		if err != nil {
			return err
		}

		// lastly insert the task information
		ref := db.Collection("tasks").NewDoc()
		err = tx.Create(ref, task)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func AddServer(ctx context.Context, db *firestore.Client, serverID int) error {
	err := db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {
		ref := db.Collection("servers").NewDoc()
		err := tx.Create(ref, entity.Server{ID: serverID})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func DeleteTask(ctx context.Context, db *firestore.Client, taskID string) error {
	err := db.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {

		// first get the ID of the server that is running this user's task
		q := db.Collection("tasks").Where("ID", "==", taskID)
		data1, err := tx.Documents(q).GetAll()
		if err != nil {
			return err
		}
		var task entity.Task
		err = data1[0].DataTo(&task)
		if err != nil {
			return err
		}

		// updates the corresponding server since its size decreased
		q2 := db.Collection("servers").Where("ID", "==", task.ServerID)
		data2, err := tx.Documents(q2).GetAll()
		if err != nil {
			return err
		}
		var server entity.Server
		err = data2[0].DataTo(&server)
		if err != nil {
			return err
		}
		dr := db.Collection("servers").Doc(data2[0].Ref.ID)
		err = tx.Set(dr, map[string]interface{}{
			"Size": server.Size - task.Size,
		}, firestore.MergeAll)
		if err != nil {
			return err
		}

		// deletes the task information from tasks
		err = tx.Delete(data1[0].Ref)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
