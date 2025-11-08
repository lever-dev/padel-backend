//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package organization

type OrganizationsRepository interface {
}
