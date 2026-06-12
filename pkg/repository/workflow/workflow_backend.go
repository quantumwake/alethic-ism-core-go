package workflow

import "github.com/quantumwake/alethic-ism-core-go/pkg/repository"

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// NewBackendFromAccess reuses an existing connection pool (so a service that
// already holds a *repository.Access doesn't open a second pool).
func NewBackendFromAccess(access *repository.Access) *BackendStorage {
	return &BackendStorage{Access: access}
}

// FindNodesByProjectID returns all workflow nodes for a project.
func (da *BackendStorage) FindNodesByProjectID(projectID string) ([]Node, error) {
	var nodes []Node
	result := da.DB.Where("project_id = ?", projectID).Find(&nodes)
	return nodes, result.Error
}

// FindEdgesByProjectID returns all workflow edges whose source or target node
// belongs to the project. workflow_edge has no project_id column, so scope by
// node membership (mirrors Python ismdb fetch_workflow_edges).
func (da *BackendStorage) FindEdgesByProjectID(projectID string) ([]Edge, error) {
	var edges []Edge
	result := da.DB.Raw(`
		SELECT * FROM workflow_edge
		WHERE source_node_id IN (SELECT node_id FROM workflow_node WHERE project_id = ?)
		   OR target_node_id IN (SELECT node_id FROM workflow_node WHERE project_id = ?)
	`, projectID, projectID).Scan(&edges)
	return edges, result.Error
}
