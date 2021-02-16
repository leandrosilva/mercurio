echo "Will try to to get notifications for the client 123 (FAIL AUTH)\n"

curl -i http://localhost:8000/api/clients/123/notifications -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjY2IiwibmFtZSI6IlVzZXIgU2l4LVNpeC1TaXgiLCJpYXQiOjE1MTYyMzkwMjJ9.nwc5Q9svmRFk2V5sCkgBqYS34sGd-lw1PJo8rwjsiYY"