echo "Will try to mark notification 1 as read for the client 123\n"

curl -i -XPUT http://localhost:8000/api/clients/123/notifications/1/read