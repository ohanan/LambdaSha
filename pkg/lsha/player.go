package lsha

type Player interface {
	DataHolder
	Order() int
	User() User
	IsAlive() bool
	Effects() Effect
}
