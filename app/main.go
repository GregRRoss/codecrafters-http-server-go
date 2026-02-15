package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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
	var echoString string
	var echoTrue bool
	var agentGet bool

	if read_result[5]==' ' {
		urlFound = true
	} else if read_result[5:10]=="echo/"{
		fmt.Println("ECHO FOUND")
		echoTrue = true
		// Echo string is characters 10 until a space
		i := 10
		for current_character := read_result[i] ; current_character!=' '; current_character = read_result[i] {
			echoString += string(current_character)
			i++
		}
		urlFound = true
	} else if read_result[5:16]=="user-agent " {
		fmt.Println("USER AGENT ENDPOINT REACHED")
		urlFound = true
		agentGet = true
	} else {
		urlFound = false
	}
	// customize response
	if urlFound {
		response += "200 " // Status Code
		response += "OK" // REASON PHRASE
		if echoTrue {
			fmt.Println(echoString)
		}
	} else {
		response += "404 " // NOT FOUND STATUS CODE
		response += "Not Found" // REASON PHRASE NOT FOUND
	}
	response += "\r\n" // CRLF for status line
	if echoTrue {
		response += "Content-Type: text/plain\r\n" // Header for format of response body
		echoLength := len(echoString)
		fmt.Print("echoLength: ")
		fmt.Println(echoLength)
		response += fmt.Sprintf("Content-Length: %d", echoLength)
		response += "\r\n" 
	} else if agentGet {
		var outputAgentName string
		cut1, cut2, _ := strings.Cut(read_result, "\r\n") // cut 1 will be the header line, cut2 is rest of header
		fmt.Println("cut1: " + cut1)
		fmt.Println("cut2: " + cut2)
		// Check each header line to find user-agent
		for {
			if len(cut1) > 11 {
				if cut1[0:12] == "User-Agent: " {
					outputAgentName = cut1[12:]
					break
				}
			}
		cut1, cut2, _ =  strings.Cut(cut2, "\r\n") // cut 1 will be the header line, cut2 is rest of header
	        fmt.Println("cut1: " + cut1)
                fmt.Println("cut2: " + cut2)	
	}
		response += outputAgentName

		
	}

	// End of Header 
	response += "\r\n" // CRLF end of headers
	
	// Body
	if echoTrue {
		response += echoString
	}
		
	fmt.Println("RESPONSE TO CLIENT")
	fmt.Println(response)
	responder.Write([]byte(response))

	// Close resources at end
	l.Close()
	responder.Close()
	// Should actually probably use defer for the above two so they still run if errors out?
}
