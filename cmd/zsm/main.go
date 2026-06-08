package main

import "github.com/khangkontum/zmx-session-manager/internal/app"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app.SetVersionInfo(version, commit, date)
	app.Main()
}
