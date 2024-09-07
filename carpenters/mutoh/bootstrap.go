package mutoh

import (
	"sudonters/zootler/carpenters/ichiro"
	"sudonters/zootler/carpenters/jiro"
	"sudonters/zootler/carpenters/saburo"
	"sudonters/zootler/carpenters/shiro"
	"sudonters/zootler/internal/app"

	"github.com/etc-sudonters/substrate/slipup"
)

type Bootstrapper struct {
	Ichiro ichiro.DataLoader
	Jiro   jiro.WorldGraph
}

func (b *Bootstrapper) Setup(z *app.Zootlr) error {
	if err := b.Ichiro.Setup(z); err != nil {
		return slipup.Describe(err, "while loading data tables")
	}
	if err := b.Jiro.Setup(z); err != nil {
		return slipup.Describe(err, "while loading logic graph")
	}
	saburo.CompileAllRules()
	shiro.CreateWorldTemplate()
	return nil
}
