package lsha

type Player interface {
	EventManager
	Account() User
	IsAlive() bool
	Effects() Effect
}
