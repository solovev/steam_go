# steam_go
> Simple steam auth util

### Installation
```
$ go get github.com/solovev/steam_go
```
### Usage
Just <code>go run main.go</code> in example dir and open [localhost:8081/login](http://localhost:8081/login) link to see how it works

Code from ./example/main.go:
```
package main

import (
	"net/http"

	"github.com/solovev/steam_go"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	opId := steam_go.NewOpenId(r)
	switch opId.Mode() {
	case "":
		http.Redirect(w, r, opId.AuthUrl(), 301)
	case "cancel":
		w.Write([]byte("Authorization cancelled"))
	default:
		steamId, err := opId.ValidateAndGetId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Do whatever you want with steam id
		w.Write([]byte(steamId))
	}
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.ListenAndServe(":8081", nil)
}

```
