package routers

import (
	"github.com/astaxie/beego"
)

func init() {
	// Serve static files for the dashboard
	beego.SetStaticPath("/", "static")
}
