package cicd

//Provider defines CI/CD provider interface
type Provider interface {
	GetTag() (string, error)
	GetActor() (string, error)
	GetOwnerAndRepo() (string, string, error)
	GetWorkDirectory() string
	GetTemplateFile() string
}
