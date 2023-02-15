package entity

//type Task struct {
//	Size int `json:"size"`
//}

type Task struct {
	ID       string `json:"ID" firestore:"ID"`
	Size     int    `json:"Size" firestore:"Size"`
	ServerID int    `firestore:"ServerID"`
}

type Server struct {
	ID   int `firestore:"ID"`
	Size int `firestore:"Size"`
}

const (
	GCP_PROJECT = "hand-made-auto-scaling"
	ZONE        = "europe-central2-b"
)
