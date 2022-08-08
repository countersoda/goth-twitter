package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/daubit/goth"
	"github.com/daubit/goth/gothic"
	"github.com/daubit/goth/providers/discord"
	"github.com/daubit/goth/providers/twitterv2"
	"github.com/gorilla/pat"
	"github.com/joho/godotenv"
)

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	goth.UseProviders(
		//twitterv2.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://127.0.0.1:3000/auth/twitterv2/callback"),
		// If you'd like to use authenticate instead of authorize in Twitter provider, use this instead.
		twitterv2.NewAuthenticate(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://127.0.0.1:3000/auth/twitterv2/callback"),
		discord.New(os.Getenv("DISCORD_KEY"), os.Getenv("DISCORD_SECRET"), "http://127.0.0.1:3000/auth/discord/callback", discord.ScopeIdentify, discord.ScopeEmail, discord.ScopeGuilds),
	)

	m := make(map[string]string)
	m["twitterv2"] = "Twitter"
	m["discord"] = "Discord"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	p := pat.New()
	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.ParseFiles("templates/success.html")
		t.Execute(res, user)
	})

	p.Get("/test/{provider}", func(res http.ResponseWriter, req *http.Request) {
		providerName, _ := gothic.GetProviderName(req)
		provider, _ := goth.GetProvider(providerName)
		value, _ := gothic.GetFromSession(providerName, req)
		sess, _ := provider.UnmarshalSession(value)
		sess.GetAuthURL()
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
		//t, _ := template.ParseFiles("templates/index.html")
		//t.Execute(res, providerIndex)
		t, _ := template.New("foo").Parse(indexTemplate)
		t.Execute(res, providerIndex)

	})

	log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", p))
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`
