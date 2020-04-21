package common

const (
	F = 1
)

type MessageEvent struct {
	VertexId *Vertex `json:"vertex"`
	Message string `json:"message"`
	Deps []*Vertex `json:"dependency,omitempty"`
}

type Vertex struct {
	Index int `json:"index"`
	Id int `json:"id"`
}
