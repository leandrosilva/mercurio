echo "Will try to broadcast some blah blah blah kind of thing to clients 123 and 456\n"

curl -i -XPOST "http://localhost:8000/api/events/broadcast" -H "Content-Type: application/json" -d'{"sourceID":"terminal","destinations":["123","456"],"data":"some blah blah blah kind of thing"}'
