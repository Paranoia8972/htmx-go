# A simple Todo app in GO

### Why?

I wanted to start with Golang and test out htmx. And I thought to myself:

> _Why not combine both?_

That's what I did!

## Getting started

### Creating the database

```bash
mkdir database
sqlite3 database/todo.db "CREATE TABLE todos (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, description TEXT, done BOOLEAN);"
```

### Running

```bash
go run main.go
```

## What's next?

- **Edit tasks:**
  Allow users to edit the title and description of existing tasks.
- **Search Functionality:**
  Add a search bar to filter tasks by title or description.
- **Categories or Tags:**
  Add the ability to categorize or tag tasks, and filter tasks by category or tag.
- **Export/Import Tasks:**
  Allow users to export their tasks to a file and import tasks from a file.
