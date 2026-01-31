package template

// Template represents a template record from the database
type Template struct {
	TemplateID      string  `gorm:"column:template_id;primaryKey"`
	TemplatePath    string  `gorm:"column:template_path"`
	TemplateContent string  `gorm:"column:template_content"`
	TemplateType    string  `gorm:"column:template_type"`
	ProjectID       *string `gorm:"column:project_id"`
}

// TableName overrides the table name used by GORM
func (Template) TableName() string {
	return "template"
}
