package forms

type ErrorSite interface {
	AddError(err error)
	NoteChildError()
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
			errs.ParentErrorSite.NoteChildError()
		}
	}
}

func (errs *SingleErrorSite) NoteChildError() {
	errs.ErrCount++
	if errs.ParentErrorSite != nil {
		errs.ParentErrorSite.NoteChildError()
	}
}

func (errs *SingleErrorSite) Invalid() bool         { return errs.ErrCount > 0 }
func (errs *SingleErrorSite) Init(parent ErrorSite) { errs.ParentErrorSite = parent }

type MultiErrorSite struct {
	Errors          []error
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
				errs.ParentErrorSite.NoteChildError()
			}
		}
	}
}

func (errs *MultiErrorSite) NoteChildError() {
	errs.ErrCount++
	if errs.ParentErrorSite != nil {
		errs.ParentErrorSite.NoteChildError()
	}
}

func (errs *MultiErrorSite) Invalid() bool         { return errs.ErrCount > 0 }
func (errs *MultiErrorSite) Init(parent ErrorSite) { errs.ParentErrorSite = parent }
