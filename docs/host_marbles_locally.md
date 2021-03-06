# Host Marbles Locally:

### <a name="runlocal"></a>Run Marbles
Lets do the very last setup for marbles.

1. Finally lets install marble's npm dependencies. Open a command prompt/terminal and browse to the root of this project.
1. In the command prompt type:
	
		> npm install gulp -g
		> npm install
		> gulp marbles1
		
1. If all goes well you should see this message in the console:

![](/doc_images/localhost1.png)

1. Go to your browser at the url specified in the console and login. You do not need to enter a password or change the prefilled username of `admin`.

![](/doc_images/localhost2.png)
	

1. Next the settuping up panel should pop up. Ideally it will walk itself through the 3 stages of initial setup.
	1. Enroll Admin - this step is communicating with your network's CA to verify the admin user credentails (enrollID/enrollSecret)
		- if it fails double check the enrollID and enrollSecret fields in your `blockchain_creds1.json` file
	1. Finding Chaincode - this step is looking for the marbles chaincode on your peer. It is using the chaincode ID found in your `blockchain_creds1.json` file. If this is a brand new network it will not exist yet. 
		- if the chaincode was instantiated but it was unable to find it try the "Retry" button.
	1. Register Marble Owners - this step will create the marble owners you specificed in the `blockchain_creds1.json` file
 
![](/doc_images/localhost3.png)

1. Once you see this message you are good to go: 

![](/doc_images/localhost4.png)
		
1. Marbles is all setup! Now [Continue tutorial 1](./tutorial_start_here.md#use).
