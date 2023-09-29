package main

import (
	"os"
	"testing"
	"webApp/pkg/repository/dbrepo"
)

var app application

// TestMain will be executed before the actual test run
func TestMain(m *testing.M) {
	//handlerのpathToTemplatesを上書き
	pathToTemplates = "./../../templates/"

	app.Session = getSession()

	app.DB = &dbrepo.TestDBRepo{}

	// it runs all of tests
	os.Exit(m.Run())
}
