
# Task Manager API

[](https://golang.org/)

A simple and efficient RESTful API for managing tasks, built with Go. The API provides endpoints for creating, retrieving, updating, and deleting tasks.

## Table of Contents

1.  [Features](#features)
2.  [Project Structure](#project-structure)
3.  [Getting Started](#getting-started)
      * [Prerequisites](#prerequisites)
      * [Installation & Setup](#installation--setup)
4.  [Running the Application](#running-the-application)
5.  [API Endpoints & Usage](#api-endpoints--usage)
      * [Create a Task](#create-a-task)
      * [Get All Tasks](#get-all-tasks)
      * [Get a Single Task](#get-a-single-task-by-title)
      * [Update a Task](#update-a-task)
      * [Delete a Task](#delete-a-task)
6.  [Testing](#testing)
7.  [Contributing](#contributing)

-----

## Features

  * **CRUD Operations**: Full support for creating, reading, updating, and deleting tasks.
  * **Robust Backend**: Built with Go, using the Gin framework for routing and GORM for database interaction.
  * **Simple Setup**: Makefile commands for easy setup and execution.
  * **PostgreSQL Database**: Uses a reliable and powerful PostgreSQL database.

-----

## Project Structure

The project follows a standard Go application layout:

```
├── backend/
│   ├── cmd/main.go         # Main application entry point
│   ├── config/routes.go    # API route definitions
│   ├── database/           # Database connection, migration
│   │   ├── db.go
│   │   └── migrate.go
│   ├── handlers/task.go    # HTTP request handlers (controllers)
|   |   ├── test/task_test.go
|   |   └── task.go
│   └── models/task.go      # GORM data models
├── go.mod                  # Go module dependencies
├── Makefile                # Commands for building and running
└── README.md               # This file
```

-----

## Getting Started

Follow these instructions to get the project up and running on your local machine.

### Prerequisites

Make sure you have the following installed:

  * **Go**: Version 1.20 or later
  * **PostgreSQL**: A running instance of the PostgreSQL server
  * **Make**: The `make` build tool

### Installation & Setup

**1. Clone the repository:**

```bash
git clone https://github.com/andyosyndoh/Task-Manager.git
cd Task-Manager
```

**2. Set up the Database:**

Connect to PostgreSQL with a superuser (e.g., `postgres`) and run the following commands to create the database and a dedicated user.

```sql
-- Create the user for the application
CREATE USER manager WITH PASSWORD 'user001';

-- Create the database
CREATE DATABASE task_manager;

-- Grant privileges to the user on the new database
GRANT ALL PRIVILEGES ON DATABASE task_manager TO manager;

-- IMPORTANT: Allow the user to create databases (for the auto-create logic in db.go)
ALTER USER manager CREATEDB;
```

**3. Configure Environment Variables:**

The application uses environment variables for database configuration. Create a `.env` file in the root of the project. You can copy the example file:

```bash
cp .env.example .env
```

Now, edit the `.env` file with your database credentials. It should look like this:

```makefile
# .env
DB_USER=manager
DB_PASSWORD=user001
```

-----

## Running the Application

To start the API server, simply use the `make` command:

```bash
# Install dependencies
make deps

# Run the programme
make run
```

The server will start on **`http://localhost:3000`**. The first time you run it, it will automatically connect to the database and migrate the `tasks` table schema.

-----

## API Endpoints & Usage

Here are examples of how to interact with the API using `curl`.

### Create a Task

  * **Endpoint**: `POST /tasks`
  * **Description**: Creates a new task.

**Request:**

```bash
curl -X POST http://localhost:3000/tasks \
-H "Content-Type: application/json" \
-d '{
      "title": "Cook",
      "description": "Some Ugali Mayai and Kachumbari itakuwa Best.",
      "status": "pending",
      "due_date": "2025-12-31T23:59:59Z"
    }'
```

**Response:**

```json
{
    "id": 1,
    "title": "Cook",
    "description": "Some Ugali Mayai and Kachumbari itakuwa Best.",
    "status": "pending",
    "due_date": "2025-12-31T23:59:59Z",
    "created_at": "2025-09-06T11:22:15.716Z",
    "updated_at": "2025-09-06T11:22:15.716Z"
}
```

### Get All Tasks

  * **Endpoint**: `GET /tasks`
  * **Description**: Retrieves a list of all tasks.

**Request:**

```bash
curl -X GET http://localhost:3000/tasks
```

**Response:**

```json
[
    {
        "id": 1,
        "title": "Cook",
        "description": "Some Ugali Mayai and Kachumbari itakuwa Best.",
        "status": "pending",
        "due_date": "2025-12-31T23:59:59Z",
        "created_at": "2025-09-06T11:22:15.716Z",
        "updated_at": "2025-09-06T11:22:15.716Z"
    },
    {
        "id": 2,
        "title": "Eat",
        "description": "Teminate anything that was cooked. Literally!",
        "status": "pending",
        "due_date": "2026-01-01T02:59:59Z",
        "created_at": "2025-09-06T11:50:19.204Z",
        "updated_at": "2025-09-06T11:50:19.204Z"
    }
]
```

### Bonus Features:

**Filtering by Status:**

* **Endpoint**: `GET /tasks?status=pending`

```bash
curl -X GET http://localhost:3000/tasks?status=pending
```

**Filtering by Due Date:**

* **Endpoint**: `GET /tasks?due_date=2025-12-31`

```bash
curl -X GET "http://localhost:3000/tasks?due_date=2025-12-31"
```

**Pagination:**

* **Endpoint**: `GET /tasks?page=1&size=1`

```bash
curl -X GET "http://localhost:3000/tasks?page=1&size=1"
```

**Search by Title:**

* **Endpoint**: `GET /tasks?search=Cook`

```bash
curl -X GET "http://localhost:3000/tasks?search=Cook"
```

### Get a Single Task by Title

  * **Endpoint**: `GET /tasks/:title`
  * **Description**: Retrieves a single task by its title.

**Request:**

```bash
curl -X GET "http://localhost:3000/tasks/Cook"
```

**Response:**

```json
{
    "id": 1,
    "title": "Cook",
    "description": "Some Ugali Mayai and Kachumbari itakuwa Best.",
    "status": "pending",
    "due_date": "2025-12-31T23:59:59Z",
    "created_at": "2025-09-06T11:22:15.716Z",
    "updated_at": "2025-09-06T11:22:15.716Z"
}
```

### Update a Task

  * **Endpoint**: `PUT /tasks/:title`
  * **Description**: Updates an existing task's description and/or status.

**Request:**

```bash
curl -X PUT http://localhost:3000/tasks/Cook \
-H "Content-Type: application/json" \
-d '{
      "description": "This is an updated description.",
      "status": "completed"
    }'
```

**Response:**

```json
{
    "id": 1,
    "title": "Cook",
    "description": "This is an updated description.",
    "status": "completed",
    "due_date": "2025-12-31T23:59:59Z",
    "created_at": "2025-09-06T11:22:15.716Z",
    "updated_at": "2025-09-06T12:37:03.551Z"
}
```

### Delete a Task

  * **Endpoint**: `DELETE /tasks/:title`
  * **Description**: Deletes a task by its title.

**Request:**

```bash
curl -X DELETE http://localhost:3000/tasks/Cook
```

**Response:**

```json
{
    "message": "Task deleted successfully"
}
```

-----

## Testing

To run the test suite, use the following command:

```bash
make test
```

-----

## Contributing

Contributions are welcome\! To contribute, please fork the repository, create a new branch, make your changes, and submit a pull request.