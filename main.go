package main

import (
	"log"

	"github.com/SwarnenduG07/wtop/ui"
)

func main() {
	dashboard := ui.NewDashboard(ui.DefaultTheme())
	if err := dashboard.Run(); err != nil {
		log.Fatalf("wtop: %v", err)
	}
}
