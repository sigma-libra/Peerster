function add_node() {

    var node = document.getElementById("new_node").value;

    $.post("/nodes", {"newNode": node});
}

function send_message() {

    var msg = document.getElementById('new_message').value;

    $.post("/message", {"newMessage": msg});

}

function send_keywords() {

    var keywords = document.getElementById('keywords').value;

    $.post('/matchingfiles', {"keywords": keywords});

}


// Use a named immediately-invoked function expression.
function getMessages() {
    $.getJSON('/message', function (data) {
        var idField = document.getElementById('MessagesField');
        idField.innerHTML = "Messages: \n" + data;
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

function getFilenames() {
    $.getJSON('/matchingfiles', function (data) { //matchingfiles
        var idField = document.getElementById('FilesField');
        idField.innerHTML = "Matching Files: \n" + data;
        setTimeout(getFilenames, 2000);
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

    $.post("/download", {"byhash": "1", "dst": dst, "hash": hash, "name": name});

}

function downloadSelectedFile(filename) {
    window.alert("Downloading "+ filename);
    $.post('/download', {"byhash": "0", "filename": filename})
}

/*
        <form>
        <label for="new_message">New Message:</label><input type="text" id="new_message"><br>
        <label for="new_node">New Node:</label><input type="text" id="new_node">
        </form>
 */