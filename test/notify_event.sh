echo "Will try to send some blah blah blah kind of thing to client 123\n"

curl -i -XPOST "http://localhost:8000/notify" -H "Content-Type: application/json" -d'{"sourceID":"terminal","clientID":"123","data":"some blah blah blah kind of thing"}'
