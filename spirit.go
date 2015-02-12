package spirit

type Spirit interface {
	Hosting(components ...Component) Spirit
	Run()
}
