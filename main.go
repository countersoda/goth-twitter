package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/pat"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitter"
)

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

func main() {
	// key := "SESSION_SECRET" // Replace with your SESSION_SECRET or similar
	// maxAge := 86400 * 30    // 30 days
	// isProd := false         // Set to true when serving over https
	// store := sessions.NewFilesystemStore("store", []byte(key))
	// store.MaxAge(maxAge)
	// store.Options.HttpOnly = true // HttpOnly should always be enabled
	// store.Options.Secure = isProd

	// gothic.Store = store
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	goth.UseProviders(
		// twitter.New("TWITTER_KEY", "TWITTER_SECRET", "http://127.0.0.1:3000/auth/twitter/callback"),
		// If you'd like to use authenticate instead of authorize in Twitter provider, use this instead.
		twitter.NewAuthenticate(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://127.0.0.1:3000/auth/twitter/callback"),
	)

	m := make(map[string]string)
	m["twitter"] = "Twitter"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{
		Providers:    keys,
		ProvidersMap: m,
	}

	p := pat.New()
	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {
		providerName, err := gothic.GetProviderName(req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		provider, err := goth.GetProvider(providerName)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		value, err := gothic.GetFromSession(providerName, req)
		fmt.Printf("User: %v, Error: %v\nSession: %v\nProvider: %v\n", nil, err, value, provider)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.ParseFiles("templates/success.html")
		t.Execute(res, nil)
	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			t, _ := template.ParseFiles("templates/success.html")
			t.Execute(res, gothUser)
		} else {
			gothic.BeginAuthHandler(res, req)
		}
	})

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(res, providerIndex)
	})
	log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", p))
}
