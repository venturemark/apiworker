package project

var (
	description = "Worker process for managing api resources asynchronously."
	gitSHA      = "n/a"
	name        = "apiworker"
	source      = "https://github.com/venturemark/apiworker"
	version     = "n/a"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
