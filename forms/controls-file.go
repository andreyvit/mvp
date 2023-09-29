package forms

import (
	"log"
	"mime/multipart"
)

type FileUpload[T comparable] struct {
	Template
	TemplateStyle
	Field
	*Binding[T]

	Required bool

	ButtonTitle string
	ButtonTag   TagOpts

	Handler func(file *multipart.FileHeader) (T, error)
}

func (FileUpload[T]) DefaultTemplate() string { return "control-file" }

func (c *FileUpload[T]) Finalize(state *State) {
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
	files := data.Files[c.FullName]
	var file *multipart.FileHeader
	if len(files) > 0 {
		file = files[0]
	}

	log.Printf("FileUpload.Process: name = %q", c.FullName)
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
