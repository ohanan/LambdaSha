package basic

import "github.com/ohanan/LambdaSha/pkg/lsha"

func Init(pb lsha.PluginBuilder) {
	pb.Name(PluginName).Version(Version).Description("this is mode for one-on-one").
		OnLoad(func(r lsha.ModeRepository) {
			r.BuildMode(initOneOnOne)
		})
}
