'use strict';
/* global process */
/* global __dirname */
/////////////////////////////////////////
///////////// Setup Node.js /////////////
/////////////////////////////////////////
var express = require('express');
var session = require('express-session');
var compression = require('compression');
var serve_static = require('serve-static');
var path = require('path');
var morgan = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var http = require('http');
var app = express();
var url = require('url');
var setup = require('./setup');
var fs = require('fs');
var cors = require('cors');

//// Set Server Parameters ////
var host = setup.SERVER.HOST;
var port = setup.SERVER.PORT;

////////  Pathing and Module Setup  ////////
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'jade');
app.engine('.html', require('jade').__express);
app.use(compression());
app.use(morgan('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded()); 
app.use(cookieParser());
app.use('/cc/summary', serve_static(path.join(__dirname, 'cc_summaries')) );												//for chaincode investigator
app.use( serve_static(path.join(__dirname, 'public'), {maxAge: '1d', setHeaders: setCustomCC}) );							//1 day cache
//app.use( serve_static(path.join(__dirname, 'public')) );
app.use(session({secret:'Somethignsomething1234!test', resave:true, saveUninitialized:true}));
function setCustomCC(res, path) {
	if (serve_static.mime.lookup(path) === 'image/jpeg')  res.setHeader('Cache-Control', 'public, max-age=2592000');		//30 days cache
	else if (serve_static.mime.lookup(path) === 'image/png') res.setHeader('Cache-Control', 'public, max-age=2592000');
	else if (serve_static.mime.lookup(path) === 'image/x-icon') res.setHeader('Cache-Control', 'public, max-age=2592000');
}
// Enable CORS preflight across the board.
app.options('*', cors());
app.use(cors());

//---------------------
// Cache Busting Hash
//---------------------
var bust_js = require('./busters_js.json');
var bust_css = require('./busters_css.json');
process.env.cachebust_js = bust_js['public/js/singlejshash'];			//i'm just making 1 hash against all js for easier jade implementation
process.env.cachebust_css = bust_css['public/css/singlecsshash'];		//i'm just making 1 hash against all css for easier jade implementation
console.log('cache busting hash js', process.env.cachebust_js, 'css', process.env.cachebust_css);


///////////  Configure Webserver  ///////////
app.use(function(req, res, next){
	var keys;
	console.log('------------------------------------------ incoming request ------------------------------------------');
	console.log('New ' + req.method + ' request for', req.url);
	req.bag = {};																			//create object for my stuff
	req.bag.session = req.session;
	
	var url_parts = url.parse(req.url, true);
	req.parameters = url_parts.query;
	keys = Object.keys(req.parameters);
	if(req.parameters && keys.length > 0) console.log({parameters: req.parameters});		//print request parameters for debug
	keys = Object.keys(req.body);
	if (req.body && keys.length > 0) console.log({body: req.body});							//print request body for debug
	next();
});

//// Router ////
app.use('/', require('./routes/site_router'));

////////////////////////////////////////////
////////////// Error Handling //////////////
////////////////////////////////////////////
app.use(function(req, res, next) {
	var err = new Error('Not Found');
	err.status = 404;
	next(err);
});
app.use(function(err, req, res, next) {														// = development error handler, print stack trace
	console.log('Error Handeler -', req.url);
	var errorCode = err.status || 500;
	res.status(errorCode);
	req.bag.error = {msg:err.stack, status:errorCode};
	if(req.bag.error.status == 404) req.bag.error.msg = 'Sorry, I cannot locate that file';
	res.render('template/error', {bag:req.bag});
});


// ============================================================================================================================
// 														Launch Webserver
// ============================================================================================================================
var server = http.createServer(app).listen(port, function() {});
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
process.env.NODE_ENV = 'production';
server.timeout = 240000;																							// Ta-da.
console.log('------------------------------------------ Server Up - ' + host + ':' + port + ' ------------------------------------------');
if(process.env.PRODUCTION) console.log('Running using Production settings');
else console.log('Running using Developer settings');


// ============================================================================================================================
// 														Deployment Tracking
// ============================================================================================================================
console.log('- Tracking Deployment');
require('cf-deployment-tracker-client').track();		//reports back to us, this helps us judge interest! feel free to remove it


// ============================================================================================================================
// ============================================================================================================================
// ============================================================================================================================
// ============================================================================================================================
// ============================================================================================================================
// ============================================================================================================================

// ============================================================================================================================
// 														Warning
// ============================================================================================================================

// ============================================================================================================================
// 														Entering
// ============================================================================================================================

// ============================================================================================================================
// 														Work Area
// ============================================================================================================================
var part1 = require('./utils/ws_part1');														//websocket message processing for part 1
var ws = require('ws');																			//websocket mod
var wss = {};
var Ibc1 = require('ibm-blockchain-js');														//rest based SDK for ibm blockchain
var ibc = new Ibc1();

// ==================================
// load peers manually or from VCAP, VCAP will overwrite hardcoded list!
// ==================================
try{
	//this hard coded list is intentionaly left here, feel free to use it when initially starting out
	//please create your own network when you are up and running
	var manual = JSON.parse(fs.readFileSync('mycreds_docker_compose.json', 'utf8'));
	var peers = manual.credentials.peers;
	console.log('loading hardcoded peers');
	var users = null;	//users are only found if security is on
	if(manual.credentials.users) users = manual.credentials.users;
	console.log('loading hardcoded users');
}
catch(e){
	console.log('Error - could not find hardcoded peers/users, this is okay if running in bluemix');
}

//filter for type1 users if we have any
function prefer_type1_users(user_array){
	var ret = [];
	for(var i in users){
		if(users[i].enrollId.indexOf('type1') >= 0) {	//gather the type1 users
			ret.push(users[i]);
		}
	}

	if(ret.length === 0) ret = user_array;				//if no users found, just use what we have
	return ret;
}

//see if peer 0 wants tls or no tls
function detect_tls_or_not(peer_array){
	var tls = false;
	if(peer_array[0] && peer_array[0].api_port_tls){
		if(!isNaN(peer_array[0].api_port_tls)) tls = true;
	}
	return tls;
}

// ==================================
// configure options for ibm-blockchain-js sdk
// ==================================
var options = {
    network: {
        peers: [peers[0]],																	//lets only use the first peer! since we really don't need any more than 1
        users: prefer_type1_users(users),													//dump the whole thing, sdk will parse for a good one
        options: {
            quiet: true, 															//detailed debug messages on/off true/false
            tls: detect_tls_or_not(peers), 											//should app to peer communication use tls?
            maxRetry: 1																//how many times should we retry register before giving up
        }
    },
    chaincode: {
        zip_url: 'https://github.com/Woda1011/marbles/archive/v2.0.zip',
        unzip_dir: 'marbles-2.0/chaincode',													//subdirectroy name of chaincode after unzipped
        git_url: 'github.com/marbles/chaincode',						//GO get http url

        //hashed cc name from prev deployment, comment me out to always deploy, uncomment me when its already deployed to skip deploying again
        deployed_name: 'fc9329be13349b6946d1e6071ca8764c57d04fbbb37629309e35c7a4943524843c6bfb920b59017b38191ea7b4c0278f8f206b672164f2cf164da5e7c9a0ad93'
    }
};

// ---- Fire off SDK ---- //
var chaincode = null;																		//sdk will populate this var in time, lets give it high scope by creating it here
ibc.load(options, function (err, cc){														//parse/load chaincode, response has chaincode functions!
	if(err != null){
		console.log('! looks like an error loading the chaincode or network, app will fail\n', err);
		if(!process.error) process.error = {type: 'load', msg: err.details};				//if it already exist, keep the last error
	}
	else{
		chaincode = cc;
		console.log('deployed chaincode:', cc);
		part1.setup(ibc, cc);																//pass the cc obj to part 1 node code

		// ---- To Deploy or Not to Deploy ---- //
		if(!cc.details.deployed_name || cc.details.deployed_name === ''){					//yes, go deploy
			cc.deploy(
			    'init',
                [],
                {delay_ms: 30000},
                function(e){ 						//delay_ms is milliseconds to wait after deploy for conatiner to start, 50sec recommended
				    check_if_deployed(e, 1);
			    }
			);
		}
		else{																				//no, already deployed
			console.log('chaincode summary file indicates chaincode has been previously deployed');
			check_if_deployed(null, 1);
		}
	}
});

//loop here, check if chaincode is up and running or not
function check_if_deployed(e, attempt){
	if(e){
		cb_deployed(e);																		//looks like an error pass it along
	}
	else if(attempt >= 15){																	//tried many times, lets give up and pass an err msg
		console.log('[preflight check]', attempt, ': failed too many times, giving up');
		var msg = 'chaincode is taking an unusually long time to start. this sounds like a network error, check peer logs';
		if(!process.error) process.error = {type: 'deploy', msg: msg};
		cb_deployed(msg);
	}
	else{
		console.log('[preflight check]', attempt, ': testing if chaincode is ready');
		chaincode.query.read(['_artefactindex'], function(err, resp){
			var cc_deployed = false;
			try{
				if(err == null){															//no errors is good, but can't trust that alone
					if(resp === 'null') cc_deployed = true;									//looks alright, brand new, no marbles yet
					else{
						var json = JSON.parse(resp);
						if(json.constructor === Array) cc_deployed = true;					//looks alright, we have marbles
					}
				}
			}
			catch(e){}																		//anything nasty goes here

			// ---- Are We Ready? ---- //
			if(!cc_deployed){
				console.log('[preflight check]', attempt, ': failed, trying again');
				setTimeout(function(){
					check_if_deployed(null, ++attempt);										//no, try again later
				}, 10000);
			}
			else{
				console.log('[preflight check]', attempt, ': success');
				cb_deployed(null);															//yes, lets go!
			}
		});
	}
}

// ============================================================================================================================
// 												WebSocket Communication Madness
// ============================================================================================================================
function cb_deployed(e){
	if(e != null){
		console.log('! looks like a deploy error, holding off on the starting the socket\n', e);
		if(!process.error) process.error = {type: 'deploy', msg: e.details};
	}
	else{
		console.log('------------------------------------------ Websocket Up ------------------------------------------');
		
		wss = new ws.Server({server: server});												//start the websocket now
		wss.on('connection', function connection(ws) {
			ws.on('message', function incoming(message) {
				console.log('received ws msg:', message);
				try{
					var data = JSON.parse(message);
					part1.process_msg(ws, data);											//pass the websocket msg to part 1 processing
				}
				catch(e){
					console.log('ws message error', e);
				}
			});
			
			ws.on('error', function(e){console.log('ws error', e);});
			ws.on('close', function(){console.log('ws closed');});
		});
		
		wss.broadcast = function broadcast(data) {											//send to all connections
			wss.clients.forEach(function each(client) {
				try{
					client.send(JSON.stringify(data));
				}
				catch(e){
					console.log('error broadcast ws', e);
				}
			});
		};
		
		// ========================================================
		// Monitor the height of the blockchain
		// ========================================================
		ibc.monitor_blockheight(function(chain_stats){										//there is a new block, lets refresh everything that has a state
			if(chain_stats && chain_stats.height){
				console.log('hey new block, lets refresh and broadcast to all', chain_stats.height-1);
				ibc.block_stats(chain_stats.height - 1, cb_blockstats);
				wss.broadcast({msg: 'reset'});
				chaincode.query.read(['_artefactindex'], cb_got_index);
			}
			
			//got the block's stats, lets send the statistics
			function cb_blockstats(e, stats){
				if(e != null) console.log('blockstats error:', e);
				else {
					chain_stats.height = chain_stats.height - 1;							//its 1 higher than actual height
					stats.height = chain_stats.height;										//copy
					wss.broadcast({msg: 'chainstats', e: e, chainstats: chain_stats, blockstats: stats});
				}
			}
			
			//got the artefact index, lets get each artefact
			function cb_got_index(e, index){
				if(e != null) console.log('artefact index error:', e);
				else{
					try{
						var json = JSON.parse(index);
						for(var i in json){
							console.log('!', i, json[i]);
							chaincode.query.read([json[i]], cb_got_artefact);					//iter over each, read their values
						}
					}
					catch(e){
						console.log('artefact index msg error:', e);
					}
				}
			}
			
			//call back for getting a artefact, lets send a message
			function cb_got_artefact(e, artefact){
				if(e != null) console.log('artefact error:', e);
				else {
					try{
						wss.broadcast({msg: 'artefact', artefact: JSON.parse(artefact)});
					}
					catch(e){
						console.log('artefact msg error', e);
					}
				}
			}
		});
	}
}