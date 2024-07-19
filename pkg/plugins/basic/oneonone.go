package basic

import "github.com/ohanan/LambdaSha/pkg/lsha"

type OneOnOne struct {
}

func buildOneOnOne() lsha.ModeBuilderV1[*OneOnOne] {
	b := lsha.NewModeBuilder[*OneOnOne](ModeOneOnOne)
	b.Description("")
	b.LimitUserCount(2, 2)
	b.OnCreate(func(e lsha.Engine, creator lsha.Player) *OneOnOne {

	})

	return b
}
