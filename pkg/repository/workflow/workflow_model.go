package workflow

import "github.com/quantumwake/alethic-ism-core-go/pkg/data"

// Node mirrors the workflow_node table: a UI canvas node, linked to a backing
// state/processor object via ObjectID. Source of truth is the Python ismdb
// workflow_storage + ismcore WorkflowNode model.
type Node struct {
	NodeID    string     `gorm:"column:node_id;type:varchar(36);primaryKey;not null" json:"node_id"`
	NodeType  string     `gorm:"column:node_type;type:varchar(64);not null" json:"node_type"`
	NodeLabel *string    `gorm:"column:node_label;type:varchar(255);null" json:"node_label,omitempty"`
	ProjectID string     `gorm:"column:project_id;type:varchar(36);not null" json:"project_id"`
	ObjectID  *string    `gorm:"column:object_id;type:varchar(36);null" json:"object_id,omitempty"`
	PositionX float64    `gorm:"column:position_x;type:double precision;not null" json:"position_x"`
	PositionY float64    `gorm:"column:position_y;type:double precision;not null" json:"position_y"`
	Width     *float64   `gorm:"column:width;type:double precision;null" json:"width,omitempty"`
	Height    *float64   `gorm:"column:height;type:double precision;null" json:"height,omitempty"`
	Metadata  *data.JSON `gorm:"column:metadata;type:jsonb;null" json:"metadata,omitempty"`
}

// TableName sets the table name for the workflow Node struct.
func (Node) TableName() string { return "workflow_node" }

// Edge mirrors the workflow_edge table: a UI canvas connection between two nodes.
// Composite primary key (source_node_id, target_node_id). The table has no
// project_id column; edges are scoped to a project via node membership.
type Edge struct {
	SourceNodeID string `gorm:"column:source_node_id;type:varchar(36);primaryKey;not null" json:"source_node_id"`
	TargetNodeID string `gorm:"column:target_node_id;type:varchar(36);primaryKey;not null" json:"target_node_id"`
	SourceHandle string `gorm:"column:source_handle;type:varchar(255)" json:"source_handle"`
	TargetHandle string `gorm:"column:target_handle;type:varchar(255)" json:"target_handle"`
	EdgeLabel    string `gorm:"column:edge_label;type:varchar(255)" json:"edge_label"`
	Type         string `gorm:"column:type;type:varchar(64)" json:"type"`
	Animated     bool   `gorm:"column:animated" json:"animated"`
}

// TableName sets the table name for the workflow Edge struct.
func (Edge) TableName() string { return "workflow_edge" }
