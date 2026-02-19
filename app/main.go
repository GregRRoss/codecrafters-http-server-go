// GregRRoss Codecrafters HTTP Server 

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"strconv"
	"flag"
	"compress/gzip" // Necessary to compress to gzip format
	"bytes" // Because gzip.NewWriter requires an io.Writer interface, and bytes provides bytes.Buffer 
	"bufio" // Necessary to manage buffers for reused requests
	"io" // Necessary for ReadFull to read exactly to end of http body
	// We could add path/filepath here, and use filepath.Join() to account for different
	// Paths in different operating systems	
)

// os.ReadFile() will return a byte array as the first return and an error as the second
// Os.Stat() can give file information and an error, and can be used to check if a file is there
// os.ReadDir() will give the directory listing

func compressStuff(stringToCompress string) []byte{
	buffer := bytes.NewBuffer(make([]byte,0,1024))
	gzipWriter := gzip.NewWriter(buffer)
	gzipWriter.Write([]byte(stringToCompress))
	gzipWriter.Close()
	return buffer.Bytes()
}

// Method to convert a list of strings to a header map
func getHeaders(headerStrings []string) map[string]string {
	headers := make(map[string]string)
	for i:=0 ; i < len(headerStrings) ; i++ {
		key, value, _ := strings.Cut(headerStrings[i], ": ")
		headers[key] = value 
	}
	return headers
}

func getBody(bufferReader *bufio.Reader, bodyLength int) string {
	fmt.Println("PARSING THE BODY")
	var body string
	perfectBuffer := make([]byte, bodyLength)
	// In the previous getheaders function, we went through the buffer and consumed everything up the double carriage return.  In this function we will do a io.FullRead to completely read the body without reading too far and going into the next http packet
	bodyBytes, _ := io.ReadFull(bufferReader, perfectBuffer)
	body = string(bodyBytes)
	return body
}

// This function handles
// from each thread
func handleConnection(responder net.Conn, fileDirectory string){
	body := "BAD BOB'S BAD BODY"
	defer responder.Close()
	connectionOpen := true 
        bufferReader := bufio.NewReader(responder)
	for connectionOpen {
	// Use readString to get each line
 	// Read Content-Length
	// use ReadFull to get exactly that content from the buffer
	stringList := make([]string, 0)
	// Assembling all the strings into one slicey
	for  {
		currentLine, _ := bufferReader.ReadString('\n')
		stringList = append(stringList, currentLine)
		// Check if the current line is just an \r\n
 		if currentLine == "\r\n" {
			break
		} 
	}
	fmt.Println("Strings pulled from request: ")
	for i:=0; i<len(stringList); i++ {
		fmt.Println(strconv.Itoa(i) + " |  " + stringList[i])
	}

	
        fmt.Println("END OF TRANSMISSION")
	headers := getHeaders(stringList)
	// Get the Content-Length to know how many bytes to go into the buffer for the body
	contentLengthString, bodyExists := headers["Content-Length"]
	contentLength := 0
	if bodyExists {
		contentLengthString = strings.TrimSpace(contentLengthString) // Gets rid of \n after content length
		contentLength, _ = strconv.Atoi(contentLengthString)
	}
	body = getBody(bufferReader, contentLength)
        // make response
        response := "HTTP/1.1 " // Status Line
        urlFound := true
	gzipCompress := false
	// Check if Accept-Encoding is in headers and get the value from it
	encoding, encoded := headers["Accept-Encoding"]
	encoding = strings.TrimSpace(encoding) // Gets rid of \r\n at end of line

	if encoding != "gzip" {
		// Encoding could be comma separated list
		encodeCut1, encodeCut2, commaExists := strings.Cut(encoding, ", ")
		for commaExists {
			encodeCut1, encodeCut2, commaExists = strings.Cut(encodeCut2, ", ")
			if encodeCut1 == "gzip"{ 
				encoding = "gzip"
				 }
		}
	}
	if encoded {
		fmt.Println("Encoding: " + encoding)
		if !(encoding == "gzip") {
			encoded = false
			fmt.Println("Invalid Encoding Detected *!*!*")
		} else {
				gzipCompress = true
		}
	} else {
		fmt.Println("No Encoding")
	}
        // Parse read bytes
                // We are looking for what comes after the GET/ and before the HTTP/1.1
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

	requestLine := stringList[0]
	// Check if we have a GET request
	if requestLine[0:3]=="GET" {
		getRequest = true
	} else if requestLine[0:4]=="POST" {
		postRequest = true
	}

	if getRequest{
		fmt.Println("GET REQUEST RECIEVED")
		if requestLine[5]==' ' { // Top level directory ping
			urlFound = true
		} else if requestLine[5:10]=="echo/"{ // ECHO endpoint
			fmt.Println("ECHO FOUND")
			echoTrue = true
			// Echo string is characters 10 until a space
			i := 10
			for current_character := requestLine[i] ; current_character!=' '; current_character = requestLine[i] {
				echoString += string(current_character)
				i++
			}
			if gzipCompress {
				echoString = string(compressStuff(echoString)) 
			}
			urlFound = true
		} else if requestLine[5:16]=="user-agent " { // USER-AGENT Endpoint
			fmt.Println("USER AGENT ENDPOINT REACHED")
			urlFound = true
			agentGet = true
		} else if requestLine[5:11]=="files/" {
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
			_, word2plus, _ := strings.Cut(requestLine, " ")
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
			if encoded {
			response += ("Content-Encoding: " + encoding + "\r\n")
			}
			echoLength := len(echoString)
			fmt.Print("echoLength: ")
			fmt.Println(echoLength)
			response += fmt.Sprintf("Content-Length: %d", echoLength)
			response += "\r\n"
		} else if agentGet {
			outputAgentName, _ = headers["User-Agent"]
			response += "Content-Type: text/plain\r\n" // Header for format of response body
			outputAgentName = strings.TrimSpace(outputAgentName) // Gets rid of \n after content length
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
		if requestLine[5:12]=="/files/" {
			fmt.Println("File POST Request")
			// Get file name
                        _, word2plus, _ := strings.Cut(requestLine, " ")
                        word2, _, _ := strings.Cut(word2plus, " ")
                        // word2 will be /files/xxxxx
                        var saveFile string = word2[7:]
			fmt.Println("New file name: " + saveFile)
			pathString := fileDirectory + "/" + saveFile
			// Parse POST request to get the file contents out of it.
			// The body of the post request is the same as the text to put in the file
			dataString := body
			os.WriteFile(pathString,[]byte(dataString),0644) //0644 is apparently the necessary permissions
			response := "HTTP/1.1 201 Created\r\n\r\n"
			responder.Write([]byte(response))
		} else {
			fmt.Println("Improper POST request")
		}
	}
	fmt.Println("AWAITING NEXT TCP")
	
} // End loop looping through multiple tcps on one connection


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
	}// End MAiIN:	
