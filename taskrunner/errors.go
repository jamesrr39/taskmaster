package taskrunner

type ErrJobNotFound struct{}

func (e *ErrJobNotFound) Error() string {
	return "Couldn't find task"
}
