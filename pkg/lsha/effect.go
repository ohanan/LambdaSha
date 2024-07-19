package lsha

type Effect interface {
	Name() string
	Description() string
}
