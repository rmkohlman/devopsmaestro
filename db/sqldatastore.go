package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

func init() {
	RegisterStore("sql", func(db Database) (DataStore, error) {
		return NewSQLDataStore(db, &SQLQueryBuilder{}), nil
	})
}

// SQLDataStore is a concrete implementation of the DataStore interface for SQL databases.
type SQLDataStore struct {
	db           Database
	queryBuilder *SQLQueryBuilder
}

// NewSQLDataStore creates a new instance of SQLDataStore
func NewSQLDataStore(db Database, queryBuilder *SQLQueryBuilder) *SQLDataStore {
	return &SQLDataStore{
		db:           db,
		queryBuilder: queryBuilder,
	}
}

// CreateProject inserts a new project into the database
func (ds *SQLDataStore) CreateProject(project *models.Project) error {
	query, values := ds.queryBuilder.BuildInsertQuery(project)
	_, err := ds.db.Execute(query, values...)
	if err != nil {
		return fmt.Errorf("failed to create project: %v", err)
	}
	return nil
}

// GetProjectByName retrieves a project by its name from the database
func (ds *SQLDataStore) GetProjectByName(name string) (*models.Project, error) {
	project := &models.Project{}
	query, values := ds.queryBuilder.BuildSelectQuery(project, "name = ?", name)

	row, err := ds.db.FetchOne(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %v", err)
	}

	if err := row.(*sql.Row).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan project: %v", err)
	}

	return project, nil
}

// UpdateProject updates an existing project in the database
func (ds *SQLDataStore) UpdateProject(project *models.Project) error {
	query, values := ds.queryBuilder.BuildUpdateQuery(project, "id = ?", project.ID)
	_, err := ds.db.Execute(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update project: %v", err)
	}
	return nil
}

// ListProjects retrieves all projects from the database
func (ds *SQLDataStore) ListProjects() ([]*models.Project, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM projects"
	rows, err := ds.db.FetchMany(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}

	defer rows.(*sql.Rows).Close()

	var projects []*models.Project
	for rows.(*sql.Rows).Next() {
		project := &models.Project{}
		if err := rows.(*sql.Rows).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, project)
	}

	if err := rows.(*sql.Rows).Err(); err != nil {
		return nil, fmt.Errorf("error iterating over project rows: %v", err)
	}

	return projects, nil
}
