package model

type WorkloadUpdate struct {
	Name            string
	Namespace       string
	Kind            string
	PreviousVersion string
	CurrentVersion  string
}
