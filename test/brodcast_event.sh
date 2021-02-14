echo "Will try to send some blah blah blah kind of thing to client 123\n"

curl -i -XPOST "http://localhost:8000/broadcast" -H "Content-Type: application/json" -d'{"sourceID":"terminal","destinationListID":["123","456"],"data":"some blah blah blah kind of thing"}'
