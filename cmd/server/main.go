package main

import (
	"fmt"
	"net/http"
    "github.com/starfederation/datastar-go/datastar"
)

func main() {
	// Serve the static HTML file on the root route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Listen for the Datastar click event
	http.HandleFunc("/interact", func(w http.ResponseWriter, r *http.Request) {

		sse := datastar.NewSSE(w,r)

		// Patches elements into the DOM.
		sse.PatchElements(
			`<div id="response-box">The Go backend says hello!</div>`
		)


	})

	fmt.Println("Server booting up on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
