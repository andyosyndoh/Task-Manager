curl -X POST http://localhost:3000/create  -H "Content-Type:
 application/json"     -d '{
        "title": "My New Task",
        "description": "This is a description for my new task.",
        "status": "pending",
        "due_date": "2025-12-31T23:59:59Z"
    }'
{"id":1,"title":"My New Task","description":"This is a description for my new ta                              sk.","status":"pending","due_date":"2025-12-31T23:59:59Z","created_at":"2025-09-                              06T11:22:15.716537237+03:00","updated_at":"2025-09-06T11:22:15.716537314+03:00"}


