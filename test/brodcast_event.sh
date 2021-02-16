echo "Will try to send some blah blah blah kind of thing to client 123\n"

curl -i -XPOST "http://localhost:8000/api/events/broadcast" -H "Content-Type: application/json" -d'{"sourceID":"terminal","destinations":["123","456"],"data":"some blah blah blah kind of thing"}'
