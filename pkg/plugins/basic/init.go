package basic

import "github.com/ohanan/LambdaSha/pkg/lsha"

func Init(pb lsha.PluginBuilder) {
	pb.Name(PluginName)
	pb.Version(1)
	pb.Description("")
	pb.OnLoad(load)
}

func load(e lsha.ModeRepository) {
	mb := e.BuildModeDef(ModeOneOnOne)
	mb.Limit(&lsha.ModeLimit{
		PlayerMinCount: 2,
		PlayerMaxCount: 2,
	})
	mb.OnStart(func(ctx lsha.Context, config any) lsha.Mode {

	})
}
