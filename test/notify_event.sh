echo "Will try to send some blah blah blah kind of thing to client 123\n"

curl -i -XPOST "http://localhost:8000/api/events/notify" -H "Content-Type: application/json" -d'{"sourceID":"terminal","destinationID":"123","data":"some blah blah blah kind of thing"}'
