package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.addIPToContext)

	// requestに対してsessionを発行する。また、requestに対してすでにsessionIDをcookieに保持していた場合にはそれをサーバー上のsession storeと照合して合致するsessionのsession dataをrのcontextに代入する。保持していなかった場合にはsessionを生成してその領域のpointerをrのcontextに代入する、そのsessionIDを脱出時にResponseWriterのcookieに登録する
	mux.Use(app.Session.LoadAndSave)

	// register routes
	mux.Get("/", app.Home)
	mux.Post("/login", app.Login)

	mux.Route("/user", func(mux chi.Router) {
		mux.Use(app.auth)
		mux.Get("/profile", app.Profile)
		mux.Post("/upload-profile-pic", app.UploadProfilePic)
	})

	// static assets
	// fileServerは指定ディレクトリに一致するファイルがあれば返してくれるHandlerを返す
	// http.FileServer→ Handlerを返すためhttp.Handleの第2引数に入れることができる
	// FileServerの引数はOpen()を持つ必要があるため、http.Dir()にpathを代入する必要がある
	// Handleの第一引数をfileserverのpathとつなぎあわせて探すためStripePrefixでpathの重複部分を削除する
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}