echo "Will try to mark notification abc as read for the client 123\n"

curl -i -XPUT http://localhost:8000/notifications/123/abc/read