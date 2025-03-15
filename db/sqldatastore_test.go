package db

import (
	"devopsmaestro/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*InMemorySQLiteDB, func()) {
	db, err := NewInMemoryTestDB()
	assert.NoError(t, err)

	// Create the project table
	_, err = db.(*InMemorySQLiteDB).conn.Exec(`
        CREATE TABLE project (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            created_at DATETIME,
            updated_at DATETIME
        )
    `)
	assert.NoError(t, err)

	// Return the database and a cleanup function
	return db.(*InMemorySQLiteDB), func() {
		db.Close()
	}
}

func TestSQLCreateProjectWithMock(t *testing.T) {
	mockDB := &MockDB{}
	dataStore := &SQLDataStore{db: mockDB, queryBuilder: &SQLQueryBuilder{}}

	project := &models.Project{Name: "Test Project", Description: "A test project"}

	// Set the expectation for the Execute method with the correct query and parameters
	mockDB.On("Execute", "INSERT INTO project (name, description) VALUES ($1, $2)", []interface{}{"Test Project", "A test project"}).Return(nil, nil)

	err := dataStore.CreateProject(project)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestSQLCreateProject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dataStore := &SQLDataStore{db: db, queryBuilder: &SQLQueryBuilder{}}

	project := &models.Project{Name: "Test Project", Description: "A test project"}

	err := dataStore.CreateProject(project)
	assert.NoError(t, err)

	// Verify the project was inserted correctly
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM project WHERE name = ?", project.Name).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLGetProjectByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dataStore := &SQLDataStore{db: db}

	projectName := "Test Project"
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()
	// Add a project to the database for testing
	_, err := db.conn.Exec("INSERT INTO project (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)", projectName, "A test project", createdAt, updatedAt)
	assert.NoError(t, err)

	project, err := dataStore.GetProjectByName(projectName)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, createdAt, project.CreatedAt)
	assert.Equal(t, updatedAt, project.UpdatedAt)
}

func TestSQLUpdateProject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dataStore := &SQLDataStore{db: db}
	project := &models.Project{
		ID:          1,
		Name:        "Test Project",
		Description: "A test project",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Insert the project into the database
	_, err := db.conn.Exec("INSERT INTO project (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		project.ID, project.Name, project.Description, project.CreatedAt, project.UpdatedAt)
	assert.NoError(t, err)

	// Update the project details
	project.Name = "Updated Project"
	project.Description = "An updated test project"
	project.UpdatedAt = time.Now().UTC()

	err = dataStore.UpdateProject(project)
	assert.NoError(t, err)

	// Verify the project was updated correctly
	var updatedProject models.Project
	err = db.conn.QueryRow("SELECT id, name, description, created_at, updated_at FROM project WHERE id = ?", project.ID).
		Scan(&updatedProject.ID, &updatedProject.Name, &updatedProject.Description, &updatedProject.CreatedAt, &updatedProject.UpdatedAt)
	assert.NoError(t, err)
	assert.Equal(t, project.Name, updatedProject.Name)
	assert.Equal(t, project.Description, updatedProject.Description)
	assert.Equal(t, project.CreatedAt, updatedProject.CreatedAt)
	assert.Equal(t, project.UpdatedAt, updatedProject.UpdatedAt)
}
