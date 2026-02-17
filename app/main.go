// GregRRoss Codecrafters HTTP Server 

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"flag"
	// We could add path/filepath here, and use filepath.Join() to account for different
	// Paths in different operating systems	
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = os.Exit
// os.ReadFile() will return a byte array as the first return and an error as the second
// Os.Stat() can give file information and an error, and can be used to check if a file is there
// os.ReadDir() will give the directory listing

// This function handles
// from each thread
func handleConnection(responder net.Conn, fileDirectory string){
	defer responder.Close()
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
        var outputAgentName string
        var echoTrue bool
        var agentGet bool
	var fileReturn bool
	var fileContentString string // Store the contents of a file that is read
	
	// We need to branch into a get versus a post request
	var getRequest bool = false
	var postRequest bool = false

	if read_result[0:3]=="GET" {
		getRequest = true
	} else if read_result[0:4]=="POST" {
		postRequest = true
	}

	if getRequest{
		fmt.Println("GET REQUEST RECIEVED")
		if read_result[5]==' ' { // Top level directory ping
			urlFound = true
		} else if read_result[5:10]=="echo/"{ // ECHO endpoint
			fmt.Println("ECHO FOUND")
			echoTrue = true
			// Echo string is characters 10 until a space
			i := 10
			for current_character := read_result[i] ; current_character!=' '; current_character = read_result[i] {
				echoString += string(current_character)
				i++
			}
			urlFound = true
		} else if read_result[5:16]=="user-agent " { // USER-AGENT Endpoint
			fmt.Println("USER AGENT ENDPOINT REACHED")
			urlFound = true
			agentGet = true
		} else if read_result[5:11]=="files/" {
			fmt.Println("FILES ENDPOINT REACHED")
			//Get rest of line to get file name
			//Check if file exists
			var fileFound bool
			// Get the files in the directory
			filesList, readError := os.ReadDir(fileDirectory)
			if readError != nil {
				fmt.Println(readError)
			}

			// Parse to get file name
			_, word2plus, _ := strings.Cut(read_result, " ")
			word2, _, _ := strings.Cut(word2plus, " ")  		
			// word2 will be /files/xxxxx	
			var getFile string = word2[7:]
			// In the future I can do strings.TrimPrefix(r.URL.Path, "/files/" 

			// Loop through the directory and see if the desired file is there.
			for fileIndex, file := range filesList {
				fileName := file.Name()
				fmt.Printf("%d - %s \n", fileIndex, fileName)
				if fileName == getFile {
					fileFound = true
				}
			} 

			if fileFound {
				pathString := fileDirectory + "/" + getFile
				fileContents, _ := os.ReadFile(pathString)
				fileContentString = string(fileContents)
				fmt.Println("Contents of " + pathString + ":")
				fmt.Println(fileContentString)
				urlFound = true
				fileReturn = true
			} else {
				urlFound = false
			}
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
			response += "Content-Type: text/plain\r\n" // Header for format of response body
			conLength := len(outputAgentName)
			fmt.Print("conLength: ")
			fmt.Println(conLength)
			response += fmt.Sprintf("Content-Length: %d", conLength)
			response += "\r\n"
			// END AGENT GET HEADER
		} else if fileReturn {  // FILE RETURN HEADER
			response += "Content-Type: application/octet-stream\r\n" // Header for format of response body
			conLength := len(fileContentString)
			fmt.Print("conLength: ")
			fmt.Println(conLength)
			response += fmt.Sprintf("Content-Length: %d", conLength)
			response += "\r\n"
		}
		// End of Header
		response += "\r\n" // CRLF end of headers

		// Body
		if echoTrue {
			response += echoString
		} else if agentGet {
			response += outputAgentName
		} else if fileReturn {
			response += fileContentString
		}

		fmt.Println("RESPONSE TO CLIENT")
		fmt.Println(response)
		responder.Write([]byte(response))
	 // END GET REQUEST IF
	} else if postRequest {
		fmt.Println("POST REQUEST RECIEVED")
		// Make sure the request looks in files
		if read_result[5:12]=="/files/" {
			fmt.Println("File POST Request")
			// Get file name
                        _, word2plus, _ := strings.Cut(read_result, " ")
                        word2, _, _ := strings.Cut(word2plus, " ")
                        // word2 will be /files/xxxxx
                        var saveFile string = word2[7:]
			fmt.Println("New file name: " + saveFile)
			pathString := fileDirectory + "/" + saveFile
			// Parse POST request to get the file contents out of it.
			 _, cut2, _ :=  strings.Cut(read_result, "\r\n\r\n") // cut 1 will be the header line, cut2 is body
			dataString := cut2
			os.WriteFile(pathString,[]byte(dataString),0644) //0644 is apparently the necessary permissions
			response := "HTTP/1.1 201 Created\r\n\r\n"
			responder.Write([]byte(response))
		} else {
			fmt.Println("Improper POST request")
		}
	}

}// End of function handling thread

func main() {

	// data := flag.String("data", "NO DATA", "Data for a POST request")
	fileDirectory := flag.String("directory", "/defaultDirectory", "Directory Path") // Get Directory for files
	flag.Parse()
		
	fmt.Println("SERVER LOG \n =================================== ")
	fmt.Println("File Directory: " + *fileDirectory)
        var connection net.Conn
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
                 if err != nil {
                        fmt.Println("Failed to bind to port 4221")
                        os.Exit(1)
                 }

	// Infinite loop of spawning threads for connections using Goroutines
	for {
	
		 connection, err = listener.Accept() // null pointer exception exists if responder is not set to l.accept
		 if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		 } 
		 go handleConnection(connection, *fileDirectory)
	
	     }
	}// End MAIN:	
