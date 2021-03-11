    package main

    import (
	    "fmt"
	    "log"
        
        app "./go"
	    transport "github.com/peramic/transport"
    )

    func main() {
        app.AddRoutes(transport.TransportRoutes)

        log.Printf("Server started")

        router := app.NewRouter()

        log.Fatal(http.ListenAndServe(":8080", router))
    }
