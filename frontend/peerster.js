
function add_node() {

    $.post("date-time.php", function (data) {

        // Display the returned data in browser

        $("#result").html(data);

    });
}

// Use a named immediately-invoked function expression.
function getMessages() {
    $.getJSON('http://127.0.0.1:8080/message', function(data) {
        // Now that we've completed the request schedule the next one.
        $("chat").empty().append("Messages: ");
        $.each(result, function (i, field) {
            $("chat").append(field);
        });
        setTimeout(getMessages, 5);
    });
}

function getNodes() {
    $.getJSON('http://127.0.0.1:8080/nodes', function(data) {
        // Now that we've completed the request schedule the next one
        $("nodes").empty().append("Nodes: ");
        $.each(result, function (i, field) {
            $("nodes").append(field + " ");
        });
        $("nodes").append("Printing")
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