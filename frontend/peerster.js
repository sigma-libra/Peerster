//Author: Sabrina Kall
function add_node() {

    var node = document.getElementById("new_node").value;

    $.post("/nodes", {"newNode": node});
}

function send_message() {

    var msg = document.getElementById('new_message').value;
    var groups = document.getElementById("new_message_groups").value;

    $.post("/message", {"newMessage": msg, "groups": groups});

}

function add_group() {
    var group = document.getElementById('new_group').value;
    $.post("/groups", {"group": group});
    //document.getElementById("GroupsField").innerHTML += "\n" + group;

}


function getGroups() {
    $.getJSON('/groups', function (data) {
        var idField = document.getElementById('GroupsField');
        idField.innerHTML = "Groups: \n" + data;
        setTimeout(getGroups, 2000);
    });
}

// Use a named immediately-invoked function expression.
function getMessages() {
    $.getJSON('/message', function (obj) {
        var selected = $("#tabs").tabs('option', 'active');
        var content = "<div id='tabs'><ul>";
        Object.keys(obj).forEach(function (key, index) {
            if (key !== "") {
                content += '<li><a href = "#tabs-' + index.toString()+ '">' + key + '</a></li>'
            }
        });

        content += "</ul>";

        Object.keys(obj).forEach(function (key, index) {
            if (key !== "") {
                content += '<pre id = "tabs-' + index.toString() + '">' +
                    '<p>' + obj[key] + '</p></pre>'
            }
            // key: the name of the object key
            // index: the ordinal position of the key within the object
        });

        content += "</div>";


        $('#MessagesField').html(content);
        $('#tabs').tabs({active:selected});


        //idField.innerHTML = "<b>Messages:</b> \n";
        //Object.keys(obj).forEach(function (key, index) {
         //   if (key !== "") {
          //      idField.innerHTML += "\n<b>" + key + ":</b>\n" + obj[key];
           // }
            // key: the name of the object key
            // index: the ordinal position of the key within the object
        //});

        setTimeout(getMessages, 2000);
    });
}

function getNodes() {
    $.getJSON('/nodes', function (data) {
        var idField = document.getElementById('PeersField');
        idField.innerHTML = "Peers: \n" + data;
        setTimeout(getNodes, 2000);
    });
}

function getId() {
    $.getJSON('/id', function (data) {
        // Now that we've completed the request schedule the next one.
        var idField = document.getElementById('PeerIdField');
        idField.innerHTML = "Peer ID: \n * Name: " + data.Name + "\n * UIPort: " + data.Port;
        document.title = "Peerster " + data.Name
    });
}

function uploadfile() {
    $.post("/uploadFile", {"dst": dst, "hash": hash, "name": name});
    return "http://localhost:8080/uploadFile"
}

function getmessageableNodes() {
    $.getJSON('/private_message', function (data) {
        // Now that we've completed the request schedule the next one.
        var messageable = document.getElementById('MessageableField');
        messageable.innerHTML = "Messageable nodes (clickeable): \n" + data;
        setTimeout(getmessageableNodes, 2000);
    });
}

//used in manually parsed html - do not delete
function openMessageWindow(e) {
    //alert($(e.target).text());
    //var t = $(e.target).text();
    var message = prompt("Enter your message for " + e, "");
    if (message != null) {
        $.post("/private_message", {"newMessage": message, "dest": e});
    }
}

function downloadFile() {

    var dst = document.getElementById('new_file_from').value;
    var hash = document.getElementById('new_file_hash').value;
    var name = document.getElementById("new_file_name").value;

    $.post("/download", {"dst": dst, "hash": hash, "name": name});

}

/*
        <form>
        <label for="new_message">New Message:</label><input type="text" id="new_message"><br>
        <label for="new_node">New Node:</label><input type="text" id="new_node">
        </form>
 */