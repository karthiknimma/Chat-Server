var net = require('net');
 
if(process.argv.length != 4){
	console.log("Usage: node %s <host> <port>", process.argv[1]);
	process.exit(1);
}

var host=process.argv[2];
var port=process.argv[3];
// var host = 'localhost'
// var port = '8000'
authenticated = false;


if(host.length >253 || port.length >5 ){
	console.log("Invalid host or port. Try again!\nUsage: node %s <port>", process.argv[1]);
	process.exit(1);
}

var client = new net.Socket();
console.log("Simple ChatClient.js developed by Karthik Nimma, SecAD");
console.log("Connecting to: %s:%s", host, port);

client.connect(port,host, connected);

function connected(){
	console.log("Connected to: %s:%s", client.remoteAddress, client.remotePort);
	console.log("You need to login before sending or receiving messages\n")
	loginsync();

}
var readlineSync = require('readline-sync');
var username;
var password;
function loginsync(){
	username = readlineSync.question('Username:');
	if(!inputValidated(username)){
		console.log("Username must have at least 5 characters. Please try again!");
		loginsync();
		return;
	}
	// handle the secret password by masking
	password = readlineSync.question('Password:',{
		hideEchoBack: true
	});
	if(!inputValidated(password)){
		console.log("Password must have atleast 5 characters.Try again!")
		loginsync();
		return
	}

	var login = '{"Username":"' + username + '","Password":"'+password+'"}';
	client.write(login)
}
function inputValidated(logindata){
	if(logindata.length >= 5){
		return true;
	}
	return false;
}

client.on("data",data => {
	console.log("\nReceived data:" + data);
	if(!authenticated){
		if(username && data.toString().includes("logged")){
		// if(username && data.includes(username+" logged")){
			console.log("You have logged in successfully with username " + username);
			authenticated = true;
			startchat();
		}else{
			console.log("Authentication failed please try again!");
			loginsync();
		}
	}

});

client.on("error",function(data) {
	console.log("Error");
	process.exit(2);
});
client.on("close",function(data){
	console.log("Connection has been disconnected");
	process.exit(3);
});

function startchat(){
	var rl = require('readline')
	var keyboard = rl.createInterface({
		input: process.stdin,
		output: process.stdout
	});
	console.log("Welcome to Chat System.Type anything to send to public chat.\n");
	console.log("Type '[To:Receiver] Message' to send to specific user.");
	console.log("Type .userlist to request latest online users.\n Type.exit to logout and close connection");

	keyboard.on('line',(input) => {
		console.log(`you typed: ${input}`);
		if(input === ".exit"){
			client.write('{"Command":"Quit"}');
			setTimeout(()=>{
				client.destroy();
				console.log("disconnected!");
				process.exit();},1);
		}else if(input === ".userlist"){
			client.write('{"Command":"UserList"}');
		}else if(input.includes("[To:")){
			endname = input.search("]");
			receiver = input.substring(4,endname);
			if((endname<0) || (receiver == "" || undefined)){
				console.log("unknown receiver.Try again!\n");
				return;
			}else{
				client.write('{"ChatType":"private","Receiver":"'+receiver+'","Message":"'+input.substring(endname+1,input.length)+'"}');
			}
		}else
			client.write('{"ChatType":"public","Message":"'+input+'"}');
	});
}

