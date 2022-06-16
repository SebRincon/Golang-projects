package HandlerForm

import (
	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
	"example.com/httpServer/handleForm"

)

type HandlerSup struct {
	gen.Supervisor
}

func (hs *HandlerSup) Init(args ...etf.Term) (gen.SupervisorSpec, error) {
	return gen.SupervisorSpec{
		Name: "handler_sup",
		Children: []gen.SupervisorChildSpec{
			gen.SupervisorChildSpec{
				Name:  "handler",
				Child: &HandleForm.Handler{},
			},
		},
		Strategy: gen.SupervisorStrategy{
			Type:      gen.SupervisorStrategySimpleOneForOne,
			Intensity: 5,
			Period:    5,
			Restart:   gen.SupervisorStrategyRestartTemporary,
		},
	}, nil
}
