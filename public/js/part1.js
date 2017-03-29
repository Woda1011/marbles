/* global new_block,formatDate, randStr, bag, $, clear_blocks, document, WebSocket, escapeHtml, window */
var ws = {};
var artefacts = new Map();


// =================================================================================
// On Load
// =================================================================================
$(document).on('ready', function() {
	connect_to_server();

	// =================================================================================
	// jQuery UI Events
	// =================================================================================
    $('#submit').click(function () {
        var fileReader = new FileReader();
        console.log('creating artefact');
        console.log('found file: ' + $('input[name="artefactFile"]').val());
        var file = $('input[name="artefactFile"]')[0].files[0];
        console.log('got file: ' + file);

        fileReader.onload = function () {
        	console.log('finished parsing as as text: ' + this.result);
        	var artefactData = JSON.parse(this.result);

        	var obj = {
                type: 'create',
                artefactType: $('select[name="artefactType"]').val(),
                artefactName: $('input[name="artefactName"]').val(),
                artefactVersion: $('input[name="artefactVersion"]').val(),
                artefact: artefactData,
                v: 1
            };

            if (obj.artefactType && obj.artefactName && obj.artefactVersion && obj.artefact) {
                console.log('creating artefact, sending', obj);
                ws.send(JSON.stringify(obj));
                showHomePanel();
            }

            return false;
        };
        console.log('Parsing file as text');
        fileReader.readAsText(file)
    });
	
	$('#homeLink').click(function(){
		showHomePanel();
	});

	$('#createLink').click(function(){
		$('input[name="name"]').val('r' + randStr(6));
	});

	//drag and drop marble
	$('#user2wrap, #user1wrap, #trashbin').sortable({connectWith: '.sortable'}).disableSelection();
	$('#user2wrap').droppable({drop:
		function( event, ui ) {
			var user = $(ui.draggable).attr('user');
			if(user.toLowerCase() != bag.setup.USER2){
				$(ui.draggable).addClass('invalid');
				transfer($(ui.draggable).attr('id'), bag.setup.USER2);
			}
		}
	});
	$('#user1wrap').droppable({drop:
		function( event, ui ) {
			var user = $(ui.draggable).attr('user');
			if(user.toLowerCase() != bag.setup.USER1){
				$(ui.draggable).addClass('invalid');
				transfer($(ui.draggable).attr('id'), bag.setup.USER1);
			}
		}
	});
	$('#trashbin').droppable({drop:
		function( event, ui ) {
			var id = $(ui.draggable).attr('id');
			if(id){
				console.log('removing marble', id);
				var obj = 	{
								type: 'remove',
								name: id,
								v: 1
							};
				ws.send(JSON.stringify(obj));
				$(ui.draggable).fadeOut();
				setTimeout(function(){
					$(ui.draggable).remove();
				}, 300);
				showHomePanel();
			}
		}
	});

    var clicked = false;
    $(document).on('click', '.ball', function(event){
        clicked = !clicked;
        showArtefactDetails(event, this.id);
    });

    $(document).on('mouseover', '.ball', function(event){
        showArtefactDetails(event, this.id);
    });

    $(document).on('mouseleave', '.marblesWrap', function(){
        if(!clicked) $('#artefactDetails').fadeOut();
    });

	// =================================================================================
	// Helper Fun
	// ================================================================================
	//show admin panel page
	function showHomePanel(){
		$('#homePanel').fadeIn(300);
		$('#createPanel').hide();
		
		var part = window.location.pathname.substring(0,3);
		window.history.pushState({},'', part + '/home');						//put it in url so we can f5
		
		//TODO Load Softwareartefacts
        console.log('getting new artefacts');
		setTimeout(function(){
			$('#user1wrap').html('');											//reset the panel
			$('#user2wrap').html('');
			ws.send(JSON.stringify({type: 'get', v: 1}));						//need to wait a bit
			ws.send(JSON.stringify({type: 'chainstats', v: 1}));
		}, 1000);
	}
	
	//transfer selected ball to user
	function transfer(marbleName, user){
		if(marbleName){
			console.log('transfering', marbleName);
			var obj = 	{
							type: 'transfer',
							name: marbleName,
							user: user,
							v: 1
						};
			ws.send(JSON.stringify(obj));
			showHomePanel();
		}
	}
});


// =================================================================================
// Socket Stuff
// =================================================================================
function connect_to_server(){
	var connected = false;

    // Redirect https requests to http so the server can handle them
    if(this.location.href.indexOf("https://") > -1) {
        this.location.href = this.location.href.replace("https://", "http://");
    }

	connect();

	function connect(){
		var wsUri = 'ws://' + document.location.hostname + ':' + document.location.port;
		console.log('Connectiong to websocket', wsUri);
		
		ws = new WebSocket(wsUri);
		ws.onopen = function(evt) { onOpen(evt); };
		ws.onclose = function(evt) { onClose(evt); };
		ws.onmessage = function(evt) { onMessage(evt); };
		ws.onerror = function(evt) { onError(evt); };
	}
	
	function onOpen(evt){
		console.log('WS CONNECTED');
		connected = true;
		clear_blocks();
        clear_artefacts();
		$('#errorNotificationPanel').fadeOut();
		ws.send(JSON.stringify({type: 'get', v:1}));
		ws.send(JSON.stringify({type: 'chainstats', v:1}));
	}

	function onClose(evt){
		console.log('WS DISCONNECTED', evt);
		connected = false;
		setTimeout(function(){ connect(); }, 5000);					//try again one more time, server restarts are quick
	}

	function onMessage(msg){
		try{
			var msgObj = JSON.parse(msg.data);
			if(msgObj.artefact){
				console.log('rec', msgObj.msg, msgObj);
				build_artefact(msgObj.artefact);
			}
			else if(msgObj.msg === 'chainstats'){
				console.log('rec', msgObj.msg, ': ledger blockheight', msgObj.chainstats.height, 'block', msgObj.blockstats.height);
				if(msgObj.blockstats && msgObj.blockstats.transactions) {
                    var e = formatDate(msgObj.blockstats.transactions[0].timestamp.seconds * 1000, '%M/%d/%Y &nbsp;%I:%m%P');
                    $('#blockdate').html('<span style="color:#fff">TIME</span>&nbsp;&nbsp;' + e + ' UTC');
                    var temp =  {
                        id: msgObj.blockstats.height,
                        blockstats: msgObj.blockstats
                    };
                    new_block(temp);								//send to blockchain.js
				}
			}
			else console.log('rec', msgObj.msg, msgObj);
		}
		catch(e){
			console.log('ERROR', e);
		}
	}

	function onError(evt){
		console.log('ERROR ', evt);
		if(!connected && bag.e == null){											//don't overwrite an error message
			$('#errorName').html('Warning');
			$('#errorNoticeText').html('Waiting on the node server to open up so we can talk to the blockchain. ');
			$('#errorNoticeText').append('This app is likely still starting up. ');
			$('#errorNoticeText').append('Check the server logs if this message does not go away in 1 minute. ');
			$('#errorNotificationPanel').fadeIn();
		}
	}
}


// =================================================================================
//	UI Building
// =================================================================================
function build_artefact(data){
	var html = '';
    addArtefact(data);
    console.log('got a artifact: ', data.artefactName);

    var color = getColor(data.artefactType);

    if (!$('#' + data.hash).length) {
        html += '<span id="' + data.hash + '" class=".artefact fa fa-circle ' + 'fa-3x' + ' ball ' + color + ' title="' + data.hash + '" user="' + bag.setup.USER2 + '"></span>';
        $('#user2wrap').append(html);
    }

    return html;
}
}

function addArtefact(artefactToAdd) {
        artefacts.set(artefactToAdd.hash, artefactToAdd);
}

function getColor(artefactType) {
	if (artefactType === 'Release') {
        return 'green';
    } else if (artefactType === 'Feature') {
        return 'blue'
    } else if (artefactType === 'Bugfix') {
        return 'red'
    } else {
        return 'white'
    }
}

function showArtefactDetails(event, id){
    var left = event.pageX - $('#artefactDetails').parent().offset().left - 50;
    if(left < 0) left = 0;

    var temp = artefacts.get(id);

    var html = '<p class="blckLegend"> Artefact Name: ' + temp.artefactName + '</p>';
    html += '<hr class="line"/><p>Created: ' + temp.timestamp + '</p>';
    html += '<p> Version: ' + temp.artefactVersion + '</p>';
    html += '<p> Type: ' + temp.artefactType + '</p>';
    html += '<p> Hash: ' + temp.hash + '</p>';
    $('#artefactDetails').html(html).css('left', left).fadeIn();
}

function clear_artefacts(){										//empty blocks
    artefact = new Map();
}