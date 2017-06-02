package codes

const (
	Nil           = iota
	BadRouting
	HashFailed
	UserExists
	WrongEmail
	WrongPassword
	BadRequest
	NotFound
	// DeleteOwner indicates that a user cannot be deleted because it is currently the owner of a class.
	DeleteOwner
)
