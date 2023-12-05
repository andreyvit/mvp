package forms

import (
	"log"
	"mime/multipart"
)

type FileUpload[T comparable] struct {
	RenderableImpl[FileUpload[T]]
	Template
	TemplateStyle
	FileField Field
	NameField Field
	Reset     Identity
	*Binding[T]

	Required bool

	ButtonTitle string
	ButtonTag   TagOpts

	Handler func(file *multipart.FileHeader) (T, error)
	Verify  func(raw string) (T, error)
}

func (FileUpload[T]) DefaultTemplate() string { return "control-file" }

func (c *FileUpload[T]) EnumFields(f func(name string, field *Field)) {
	f("", &c.FileField)
	f("filename", &c.NameField)
}

func (c *FileUpload[T]) Finalize(state *State) {
	state.AssignSubidentity("remove", &c.Reset)
	c.Binding.Validate(func(value T) (T, error) {
		if c.Required {
			var zero T
			if value == zero {
				return value, ErrRequired
			}
		}
		return value, nil
	})
}

func (c *FileUpload[T]) Process(data *FormData) {
	if data.Action == c.Reset.FullName {
		var zero T
		c.Binding.Set(zero)
	} else if c.NameField.RawFormValue == "" {
		var zero T
		c.Binding.Set(zero)
	} else if c.NameField.RawFormValue != "" && c.Verify != nil {
		v, err := c.Verify(c.NameField.RawFormValue)
		if err != nil {
			c.Binding.ErrSite.AddError(err)
		} else {
			c.Binding.Set(v)
		}
	}

	files := data.Files[c.FileField.FullName]
	var file *multipart.FileHeader
	if len(files) > 0 {
		file = files[0]
	}

	// log.Printf("FileUpload.Process: name = %q", c.FileField.FullName)
	if file == nil {
		return
	}

	value, err := c.Handler(file)
	if err != nil {
		log.Printf("FileUpload.Process: err = %v", err)
		c.ErrSite.AddError(err)
		return
	}

	log.Printf("FileUpload.Process: value = %v", value)
	c.Binding.Set(value)
}
