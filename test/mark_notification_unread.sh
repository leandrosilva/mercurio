echo "Will try to mark notification 1 as read for the client 123\n"

curl -i -XPUT http://localhost:8000/api/clients/123/notifications/1/unread -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwibmFtZSI6IlVzZXIgT25lLVR3by1UaHJlZSIsImlhdCI6MTUxNjIzOTAyMn0.VOCXaMknw5CG002hh1iSZGvdI-ug76F1GbSYRVu_RuU"