package AppForm

import (
	"fmt"

	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
	"example.com/httpServer/handlerForm"

)

type App struct {
	gen.Application
}

var (
	Handler_sup = &HandlerForm.HandlerSup{}
)

func (a *App) Load(args ...etf.Term) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "WebApp",
		Description: "Demo Web Application",
		Version:     "v.1.0",
		Children: []gen.ApplicationChildSpec{
			gen.ApplicationChildSpec{
				Child: Handler_sup,
				Name:  "handler_sup",
			},
		},
	}, nil
}

func (a *App) Start(process gen.Process, args ...etf.Term) {
	fmt.Println("Application started!")
}
