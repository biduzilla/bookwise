package models

type ModelInterface[D any] interface {
	ToDTO() *D
}

type DTOInterface[T any] interface {
	ToModel() *T
}
