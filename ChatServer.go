/* Simple ChatServer in GoLang by Karthik Nimma for SecAD*/
package main

import (
	"fmt"
	"net"
	"os"
	"encoding/json"

)

const BUFFERSIZE int = 1024

type User struct {
	Username string
	Login bool
	Key string
}

type Command struct{
	Command string //userlist,quit,etc
}

type ChatMessage struct {
	ChatType string //private or public
	Message string
	Receiver string //only available in private chat
}

var authenticated_clients = make(map[net.Conn]string)	//map[keytype]valuetype
var lostclient = make(chan net.Conn)
var newclient = make(chan net.Conn)

// To implement yet..............................
// var currentLoggedUser User
var allLoggedIn_conns = make(map[net.Conn]interface{})
var usersList = make(map[string]bool)
var message = make(chan string)
var activeUser string
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <port>\n", os.Args[0])
		os.Exit(0)
	}
	port := os.Args[1]
	if len(port) > 5 {
		fmt.Println("Invalid port value. Try again!")
		os.Exit(1)
	}
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Cannot listen on port '" + port + "'!\n")
		os.Exit(2)
	}
	fmt.Println("ChatServer in GoLang developed by Karthik Nimma")
	fmt.Printf("ChatServer is listening on port '%s' ...\n", port)
	
	//newclient := make(chan net.Conn)
	go func(){
		for{
			client_conn, _ := server.Accept()	
			go login(client_conn)
		}	
	}()
	for{
		select{
		case client_conn := <- newclient:
			go authenticating(client_conn)
		case client_conn := <- lostclient:
			delete(authenticated_clients,client_conn)
			delete(allLoggedIn_conns,client_conn)
			byemessage := fmt.Sprintf("client '%s' is Disconnected!\n# of connected clients: %d\n",client_conn.RemoteAddr().String(),len(authenticated_clients))
			go sendtoAll([]byte(byemessage))
		}
	}
}
func is_authenticated(client_conn net.Conn) bool {
	if allLoggedIn_conns[client_conn] != nil {
		return true
	}
	return false
}

func authenticating(client_conn net.Conn){
	authenticated_clients[client_conn] = client_conn.RemoteAddr().String()
	var newuser User
	newuser.Key = client_conn.RemoteAddr().String()
	newuser.Username = activeUser
	newuser.Login = true
	allLoggedIn_conns[client_conn] = newuser
	sendto(client_conn,[]byte("You are authenticated.Welcome to chat system!\n"))
	
	// welcomemessage := //combined the two current messages with Sprintf
	// 	fmt.Sprintf("A new client '%s' connected! \n # of connected clients :%d\n",client_conn.RemoteAddr().String(),len(authenticated_clients))
	msg := getUserlist()
	myString := string(msg[:])	
	welcomemessage := //combined the two current messages with Sprintf
	fmt.Sprintf("\nNew User '%s' logged in to Chat System from %s.Online users:'%s'.Total %d connections\n",newuser.Username,client_conn.RemoteAddr().String(),myString,len(authenticated_clients))	

	fmt.Println(welcomemessage)
	go sendtoAll([]byte (welcomemessage))	
	go client_goroutine(client_conn)
}
func client_goroutine(client_conn net.Conn){
	
	var buffer [BUFFERSIZE]byte
	go func(){
		for {
			byte_received, read_err := client_conn.Read(buffer[0:])
			if read_err != nil {
				fmt.Println("Error in receiving...")
				lostclient <- client_conn
				return
			}
			// go sendtoAll(buffer[0:byte_received])
			fmt.Printf("Received data: %s from '%s'\n",
				buffer[0:byte_received], client_conn.RemoteAddr().String())
				data := decrypt(client_conn, buffer[0:byte_received])
				handleMessages(client_conn,data)
		}
	}()		
}
func decrypt(client_conn net.Conn, data []byte) []byte {
	return data
}
func encrypt(client_conn net.Conn, data []byte) []byte {
	return data
}


func sendtoAll(data []byte) {
	for client_conn,_ := range authenticated_clients{
		_, write_err := client_conn.Write(data)
		if write_err != nil{
			fmt.Println("Error in sending...")
			continue;
		}
	}
	fmt.Printf("Send data: %s to all clients ! \n",data)
}
func sendto(client_conn net.Conn,data []byte){
	_, write_err := client_conn.Write(data)
	if write_err != nil{
		fmt.Println("Error in sending...")
		return
	}
}

// Assignment 1 authentication
func login(client_conn net.Conn){

	fmt.Printf("Client is connected: %s. Waiting for authentication! ",client_conn.RemoteAddr().String())
	var buffer [BUFFERSIZE]byte
	byte_received, read_err := client_conn.Read(buffer[0:])
	if read_err != nil {
		fmt.Println("Error in receiving...")
		lostclient <- client_conn
		return
	}
	clientdata := buffer[0:byte_received]
	fmt.Printf("Received data: %s,len=%d \n", clientdata, len(clientdata))
		
	// if checklogin fails dont add to client list
	checklogin, username, msg2 := checklogin(clientdata)



	if(!checklogin){
		// send error message back to client
		activeUser = ""
		//fmt.Println("DEBUG> Invalid JSON login format")
		sendto(client_conn,[]byte(msg2))
		login(client_conn)
	}else{
		// add client to the connected client lists
		fmt.Printf("Valid json format Valid username and password. Username =" + username + ". Message=" +msg2)
		activeUser = username
		authentication_msg := username+" "+msg2
		sendto(client_conn,[]byte(authentication_msg))
		newclient <- client_conn

		//TRIED TO IMPLEMENT FUNCTIONALITY TO ALLOW USERS TO LOGIN FROM MULTIPLE DEVICES WITHOUT CAUSING REDUNDANT USERNAMES
		// IN USERLIST.
		//if searchUser(client_conn,activeUser) {
		//	go client_goroutine(client_conn)
		//}else{
		//	newclient <- client_conn
		//}
	}
}
func searchUser(client_conn net.Conn,currentUser string) bool{
	for client_conn,_ := range allLoggedIn_conns{
		user := allLoggedIn_conns[client_conn].(User)
		if(user.Username == currentUser){
			return true
		}
	}
	return false
}
func checklogin(data []byte) (bool,string,string){
	type Account struct{
		Username string
		Password string
	}
	var account Account
	err := json.Unmarshal(data, &account)
	if err!=nil || account.Username == "" || account.Password == ""{
		fmt.Printf("JSON parsing error: %s\n",err)
		return false,"",`[BAD LOGIN] Expected: {"Username":"..","Password":".."}`
	}
	fmt.Printf("DEBUG>GOT: account=%s\n", account)
	fmt.Printf("DEBUG>Got: username=%s,password=%s\n",account.Username,account.Password)

	accountExists := checkaccount(account.Username,account.Password)
	if accountExists{
		fmt.Println("DEBUG> Username and password found!")
		return true,account.Username,"logged"
	}
	fmt.Println("DEBUG> Invalid username or password")
	return false,"","Invalid username or password\n"
}
func checkaccount(name string,passkey string) (bool){
	if(string(name) == "karthik" && string(passkey) == "jellybean"){
		return true
	}
	if(string(name) == "karthik1" && string(passkey) == "jellybean"){
		return true
	}
	if(string(name) == "karthik2" && string(passkey) == "jellybean"){
		return true
	}	
	if(string(name) == "testuser" && string(passkey) == "testpassword"){
		return true
	}		
	return false

}
func handleMessages(client_conn net.Conn, data []byte){
	if !is_authenticated(client_conn){
		sendto(client_conn,[]byte("You havent been authenticated.Message ignored!"))
		return
	}
	var command Command	
	command_err := json.Unmarshal(data, &command)
	if command_err != nil ||  command.Command == ""{
		var chatMessage ChatMessage
		chat_err := json.Unmarshal(data, &chatMessage)
		if chat_err != nil || chatMessage.ChatType == "" {
			fmt.Printf("Unknown data type = %s\n", data)
			errorMessage(client_conn)
			return
		}
		// Chat message
		fmt.Printf("DEBUG> chat message = %s\n", chatMessage)
		if chatMessage.ChatType == "private" {
			privateChat(client_conn,chatMessage)
		} else if chatMessage.ChatType=="public"{
			publicChat(client_conn,chatMessage)
		}
	}else{
		commandResponse(client_conn, command)
	}
}	

func commandResponse(client_conn net.Conn, command Command){
	fmt.Printf("DEBUG> Command = %s\n", command.Command)
	switch command.Command{
		case "UserList":
			fmt.Printf("DEBUG>get Userlist and return\n")
			userList := getUserlist()
			fmt.Printf("DEBUG> userListJSON: %s\n", userList)
			sendto(client_conn,userList)
		case "Quit":
			fmt.Printf("DEBUG> Client quited!\n")
			lostclient <- client_conn
	}
}
func errorMessage(client_conn net.Conn){
	sendto(client_conn, []byte("[BAD DATA: Unknown command or Data format "))
	expected_format := fmt.Sprintf(`Expected format: {"Command":"Quit"} | {"Command":"Userlist"} | {"ChatType":"private","Receiver":"..","Message":".."}`)
	sendto(client_conn,[]byte(expected_format))
}
func privateChat(client_conn net.Conn, chatMessage ChatMessage){
	fmt.Printf("DEBUG> Private chat to: %s. Message: %s\n", chatMessage.Receiver, chatMessage.Message)
	if allLoggedIn_conns[client_conn] != nil{
		user := allLoggedIn_conns[client_conn].(User)
		message := fmt.Sprintf("Private message from '"+user.Username+"':%s",chatMessage.Message)
		for receiver_client_conn , _ := range allLoggedIn_conns {
			if allLoggedIn_conns[receiver_client_conn] != nil {
				receiverUser := allLoggedIn_conns[receiver_client_conn].(User)
				// one user maybe logged in from different devices
				if receiverUser.Username == chatMessage.Receiver{
					sendto(receiver_client_conn, []byte(message))
				}
			}
		}
	}
}

func publicChat(client_conn net.Conn, chatMessage ChatMessage){
	fmt.Printf("DEBUG> Public chat.Message: %s\n", chatMessage.Message)
	if allLoggedIn_conns[client_conn] != nil{
		user := allLoggedIn_conns[client_conn].(User)
		message := fmt.Sprintf("Public Message from '"+user.Username+"':%s", chatMessage.Message)
		sendtoAll([]byte(message))
	}
}
func getUserlist() []byte{
	//userList := []string{}
	var userList []string
	debug_len :=0
	for _, usereach := range allLoggedIn_conns {
		debug_len = debug_len + 1
		var usereach1 User = usereach.(User)
		fmt.Printf("DEBUG> getUserlist() -> i=%d user Username = %s\n",debug_len,usereach1.Username)
		userList = append(userList," "+usereach1.Username)
	}
   // store to byte array
	// var str = []string{"str1","str2"}
	var x = []byte{}

	for i:=0; i<len(userList); i++{
	    b := []byte(userList[i])
	    for j:=0; j<len(b); j++{
	        x = append(x,b[j])
	    }
	}
	return x
}
