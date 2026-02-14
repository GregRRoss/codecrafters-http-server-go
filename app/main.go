package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
        var responder net.Conn
	// TODO: Uncomment the code below to pass the first stage
	//
	 l, err := net.Listen("tcp", "0.0.0.0:4221")
	 if err != nil {
	 	fmt.Println("Failed to bind to port 4221")
	 	os.Exit(1)
	 }
	
	 responder, err = l.Accept() // null pointer exception exists if responder is not set to l.accept
	 if err != nil {
	 	fmt.Println("Error accepting connection: ", err.Error())
	 	os.Exit(1)
	 }
	// Read get request
	buffer := make([]byte, 1024) // Create 'Slice' i.e. arrayList needed for net pkg
	var number_of_bytes_read int
	number_of_bytes_read, _ = responder.Read(buffer)
	read_result := string(buffer[:number_of_bytes_read])
 	fmt.Println("START OF TRANSMISSION")	
	fmt.Println(read_result)	
	fmt.Println("END OF TRANSMISSION")
	// make response
	response := "HTTP/1.1 " // Status Line
	urlFound := true

	// Parse read bytes
        	// We are looking for what comes after the GET / and before the HTTP/1.1	
		// Slice we want starts at 01234... character 5 must be a space
	if read_result[5]==' ' {
		urlFound = true
	} else {
		urlFound = false
	}
	// customize response
	if urlFound {
		response += "200 " // Status Code
		response += "OK" // REASON PHRASE
	} else {
		response += "404 " // NOT FOUND STATUS CODE
		response += "Not Found" // REASON PHRASE NOT FOUND	
	}

	// End response 
        response += "\r\n" // CRLF end of status line
	response += "\r\n" // CRLF end of headers
	responder.Write([]byte(response))

	// Close resources at end
	l.Close()
	responder.Close()
	// Should actually probably use defer for the above two so they still run if errors out?
}
