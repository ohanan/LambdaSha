package lsha

type Player interface {
	Account() User
	IsAlive() bool
	Effects() Effect
}
