echo "Will try to send some blah blah blah kind of thing to client 123\n"

curl -i -XPOST "http://localhost:8000/api/events/unicast" -H "Content-Type: application/json" -d'{"sourceID":"terminal","destinationID":"123","data":"some blah blah blah kind of thing"}' -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjY2IiwibmFtZSI6IlVzZXIgU2l4LVNpeC1TaXgiLCJpYXQiOjE1MTYyMzkwMjJ9.nwc5Q9svmRFk2V5sCkgBqYS34sGd-lw1PJo8rwjsiYY"
