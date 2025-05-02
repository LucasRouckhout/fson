// A simple example of how you can use fson to write out a JSON response the context of an HTTP endpoint
// In this case the performance gains you would get from avoiding the allocations are not that big, probably negligible,
// but they can become significant if scaled up to many request per second and bigger JSON bodies.

package main

import (
	"github.com/LucasRouckhout/fson"
	"log"
	"net/http"
	"sync"
)

var buffPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 100)
	},
}

// User is a simple struct that we want to encode into JSON
type User struct {
	Name  string
	Email string
}

func main() {
	// Some kind of user endpoint
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		// You would have some logic to fetch a user here I just hard code one
		user := User{
			Name:  "John Doe",
			Email: "johndoe@example.com",
		}

		// Get yourself a buffer
		buff := buffPool.Get().([]byte) // Avoid allocations by taking from a pool
		defer buffPool.Put(buff)

		// Encode the struct into JSON
		b := fson.NewObject(buff).
			String("name", user.Name).
			String("email", user.Email).
			Build()

		// Write out the headers followed by the JSON body
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
