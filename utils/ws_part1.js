// ==================================
// Part 1 - incoming messages, look for type
// ==================================
var ibc = {};
var chaincode = {};
var async = require('async');
var hash = require('object-hash');

module.exports.setup = function(sdk, cc){
	ibc = sdk;
	chaincode = cc;
};

module.exports.process_msg = function(ws, data){
	if(data.v === 1){																						//only look at messages for part 1
		if(data.type == 'create'){
			console.log('its a create!');

            data.artefactHash = hash(data);
            console.log('data hashed with sha1: ' + JSON.stringify(data));

			if(data.artefactType && data.artefactHash && data.artefactName && data.artefactVersion && data.artefact){
				chaincode.invoke.init_artefact([data.artefactVersion, data.artefactName, data.artefactHash, data.artefactType, JSON.stringify(data.artefact)], cb_invoked);
			}
		}
		//Gets called everytime the homepanel is loaded
		else if(data.type == 'get'){
			console.log('get artefact msg');
			chaincode.query.read(['_artefactindex'], cb_got_index);
            chaincode.query.read(['_deviceIndex'], cb_gotDeviceIndex);
		}
		else if(data.type == 'deploy'){
			console.log('deploying msg');
			if(data.id && data.hash){
				chaincode.invoke.deploy_artifact([data.id, data.hash], cb_invoked);
			}
		}
		else if(data.type == 'remove'){
			console.log('removing msg');
			if(data.name){
				chaincode.invoke.delete([data.name]);
			}
		}
		else if(data.type == 'chainstats'){
		    //TODO Show already created Artefacts
			console.log('chainstats msg');
			ibc.chain_stats(cb_chainstats);
		}
	}

	//got the artefact index, lets get each artefact
	function cb_got_index(e, index){
		if(e != null) console.log('[ws error] did not get artefact index:', e);
		else{
			try{
				var json = JSON.parse(index);
                console.log('[ws info] got artefact index:', index);
				var keys = Object.keys(json);
				var concurrency = 1;

				//serialized version
				async.eachLimit(keys, concurrency, function(key, cb) {
					console.log('!', json[key]);
					chaincode.query.read([json[key]], function(e, artefact) {
						if(e != null) console.log('[ws error] did not get artefact:', e);
						else {
							if(artefact) sendMsg({msg: 'artefact', e: e, artefact: JSON.parse(artefact)});
							cb(null);
						}
					});
				}, function() {
					sendMsg({msg: 'action', e: e, status: 'finished'});
                    //console.log('Start to read Devices...');
					//chaincode.query.read(['_deviceIndex'], cb_gotDeviceIndex);
				});
			}
			catch(e){
				console.log('[ws error] could not parse response', e);
			}
		}
	}

    function cb_gotDeviceIndex(e, index) {
        if (e != null) {
            console.log('[ws error] did not get device index:', e);
        }
        else {
            try {
                var json = JSON.parse(index);
                console.log('[ws info] got device index:', index);
                var keys = Object.keys(json);
                var concurrency = 1;

                async.eachLimit(keys, concurrency, function (key, cb) {
                    console.log('!', json[key]);
                    chaincode.query.read([json[key]], function (e, device) {
                        if (e != null) {
                            console.log('[ws error] did not get device:', e);
                        }
                        else if (device){
                            device = JSON.parse(device);
                            chaincode.query.read([device.currentArtifactHash], function (e, artifact) {
                                if (e !== null) {
                                    console.log('[ws error] did not get artifact for device:', e);
                                }
                                else if (artifact) {
                                    sendMsg({msg: 'device', e: e, device: device, currentArtifact: JSON.parse(artifact)});
                                }
                            });
                            sendMsg({msg: 'device', e: e, device: device});
                        }
                        cb(null);
                    });
                }, function () {
                    sendMsg({msg: 'action', e: e, status: 'finished'});
                });
            }
            catch (e) {
                console.log('[ws error] could not parse response', e);
            }
        }
    }
	
	function cb_invoked(e, a){
		console.log('response: ', e, a);
	}
	
	//call back for getting the blockchain stats, lets get the block stats now
	function cb_chainstats(e, chain_stats){
		if(chain_stats && chain_stats.height){
			chain_stats.height = chain_stats.height - 1;								//its 1 higher than actual height
			var list = [];
			for(var i = chain_stats.height; i >= 1; i--){								//create a list of heights we need
				list.push(i);
				if(list.length >= 8) break;
			}
			list.reverse();																//flip it so order is correct in UI
			async.eachLimit(list, 1, function(block_height, cb) {						//iter through each one, and send it
				ibc.block_stats(block_height, function(e, stats){
					if(e == null){
						stats.height = block_height;
						sendMsg({msg: 'chainstats', e: e, chainstats: chain_stats, blockstats: stats});
					}
					cb(null);
				});
			}, function() {
			});
		}
	}
	
	//send a message, socket might be closed...
	function sendMsg(json){
		if(ws){
			try{
				ws.send(JSON.stringify(json));
			}
			catch(e){
				console.log('[ws error] could not send msg', e);
			}
		}
	}
};