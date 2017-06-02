package codes

const (
	// Nil indicates an unknown error. Should you encounter a Nil status, fret not! It will be reported to the server admin.
	Nil = iota
	// Unauthenticated indicates that no authorization token was passed to the service, or it was invalid or malformed.
	Unauthenticated
	// NotFound indicates that the requested resource could not be found, or the user is not allowed to view it.
	NotFound
	// Forbidden indicates that the user is not allowed to perform an action.
	Forbidden
	// MustSetOwner indicates that the user may not leave a class until they have resigned as next owner, setting somebody else to replace them.
	MustSetOwner
	// UserEnrolled indicates that the user is already enrolled in a class, and thus cannot re-enroll.
	UserEnrolled
	// BadRequest indicates that the request is malformed or invalid.
	BadRequest
)
