echo "Will try to to get notification 1 for the client 123\n"

curl -i http://localhost:8000/api/clients/123/notifications/1 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwibmFtZSI6IlVzZXIgT25lLVR3by1UaHJlZSIsImlhdCI6MTUxNjIzOTAyMn0.VOCXaMknw5CG002hh1iSZGvdI-ug76F1GbSYRVu_RuU"