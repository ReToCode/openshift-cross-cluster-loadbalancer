package models

const (
	HOST_LIST = "hostList"
)

type BaseModel struct {
	Mutation string `json:"mutation"`
	Message  interface{}
}
