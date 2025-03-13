package main

import "github.com/tommjj/gotcpio"

func main() {
	server := gotcpio.NewServer(":8080")

	server.On("connection", func(c *gotcpio.Conn) {

		c.On("message", func(data gotcpio.Data) {
			c.Emit("message", data)
		})

		c.Emit("message", []byte("Hello, World!\n"))
	})

	server.ListenAndServe()
}
