package forms

type ErrorSite interface {
	AddError(err error)
	NoteChildError(err error)
	Init(parent ErrorSite)
}

type SingleErrorSite struct {
	Error           error
	ParentErrorSite ErrorSite
	ErrCount        int
}

func (errs *SingleErrorSite) AddError(err error) {
	if err != nil && errs.Error == nil {
		errs.Error = err
		errs.ErrCount++
		if errs.ParentErrorSite != nil {
			errs.ParentErrorSite.NoteChildError(err)
		}
	}
}

func (errs *SingleErrorSite) NoteChildError(err error) {
	errs.ErrCount++
	if errs.ParentErrorSite != nil {
		errs.ParentErrorSite.NoteChildError(err)
	}
}

func (errs *SingleErrorSite) ErrorStr() string {
	return ErrorStr(errs.Error)
}

func (errs *SingleErrorSite) Invalid() bool         { return errs.ErrCount > 0 }
func (errs *SingleErrorSite) Init(parent ErrorSite) { errs.ParentErrorSite = parent }

type MultiErrorSite struct {
	Errors          []error
	ChildErrors     []error
	ParentErrorSite ErrorSite
	ErrCount        int
	CaptureErrors   bool
}

func (errs *MultiErrorSite) AddError(err error) {
	if err != nil {
		errs.ErrCount++
		if !errs.CaptureErrors && errs.ParentErrorSite != nil {
			errs.ParentErrorSite.AddError(err)
		} else {
			errs.Errors = append(errs.Errors, err)
			if errs.ParentErrorSite != nil {
				errs.ParentErrorSite.NoteChildError(err)
			}
		}
	}
}

func (errs *MultiErrorSite) NoteChildError(err error) {
	errs.ErrCount++
	errs.ChildErrors = append(errs.ChildErrors, err)
	if errs.ParentErrorSite != nil {
		errs.ParentErrorSite.NoteChildError(err)
	}
}

func (errs *MultiErrorSite) Invalid() bool         { return errs.ErrCount > 0 }
func (errs *MultiErrorSite) Init(parent ErrorSite) { errs.ParentErrorSite = parent }

func ErrorStr(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(interface{ PublicError() string }); ok {
		s := e.PublicError()
		if s != "" {
			return s
		}
	}
	return err.Error()
}
