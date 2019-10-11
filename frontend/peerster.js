
function add_node() {

    var node = document.getElementById("new_node").value;

    $.post("/nodes", {"newNode":node});
}

function send_message() {

    var msg = document.getElementById('new_message').value;

    $.post("/message", {"newMessage":msg});


}

// Use a named immediately-invoked function expression.
function getMessages() {
    $.getJSON('http://127.0.0.1:8080/message', function(data) {
        var idField = document.getElementById('MessagesField');
        idField.innerHTML = "Messages: \n" + data;
        setTimeout(getMessages, 5000);
    });
}

function getNodes() {
    $.getJSON('http://127.0.0.1:8080/nodes', function(data) {
        var idField = document.getElementById('PeersField');
        idField.innerHTML = "Peers: \n" + data;
        setTimeout(getNodes, 5000);
    });
}

function getId() {
    $.getJSON('http://127.0.0.1:8080/id', function(data) {
        // Now that we've completed the request schedule the next one.
        var idField = document.getElementById('PeerIdField');
        idField.innerHTML = "Peer ID: " + data;
    });
}

/*
        <form>
        <label for="new_message">New Message:</label><input type="text" id="new_message"><br>
        <label for="new_node">New Node:</label><input type="text" id="new_node">
        </form>
 */