curl -X POST http://localhost:3000/tasks  -H "Content-Type:
 application/json"     -d '{
        "title": "My New Task",
        "description": "This is a description for my new task.",
        "status": "pending",
        "due_date": "2025-12-31T23:59:59Z"
    }'
{"id":1,"title":"My New Task","description":"This is a description for my new ta                              sk.","status":"pending","due_date":"2025-12-31T23:59:59Z","created_at":"2025-09-                              06T11:22:15.716537237+03:00","updated_at":"2025-09-06T11:22:15.716537314+03:00"}


 curl -X GET "http://localhost:3000/tasks/Task"
{"id":2,"title":"Task","description":"This is a description for my new task.","status":"pending","due_date":"2026-01-01T02:59:59+03:00","created_at":"2025-09-06T11:50:19.204726+03:00","updated_at":"2025-09-06T11:50:19.204726+03:00"}

curl -X GET "http://localhost:3000/tasks"
[{"id":1,"title":"My New Task","description":"This is a description for my new task.","status":"pending","due_date":"2026-01-01T02:59:59+03:00","created_at":"2025-09-06T11:22:15.716537+03:00","updated_at":"2025-09-06T11:22:15.716537+03:00"},{"id":2,"title":"Task","description":"This is a description for my new task.","status":"pending","due_date":"2026-01-01T02:59:59+03:00","created_at":"2025-09-06T11:50:19.204726+03:00","updated_at":"2025-09-06T11:50:19.204726+03:00"}]

curl -X PUT http://localhost:3000/tasks/Task \
    -H "Content-Type: application/json" \
    -d '{
        "description": "This is an updated description for SimpleTask.",
        "status": "completed"
    }'
{"id":2,"title":"Task","description":"This is an updated description for SimpleTask.","status":"completed","due_date":"2026-01-01T02:59:59+03:00","created_at":"2025-09-06T11:50:19.204726+03:00","updated_at":"2025-09-06T12:37:03.551790077+03:00"}

curl -X DELETE http://localhost:3000/tasks/Eating
{"message":"Task deleted successfully"}