package forms

import (
	"log"
	"mime/multipart"
)

type FileUpload[T any] struct {
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
}

func (c *FileUpload[T]) Process(data *FormData) {
	files := data.Files[c.FullName]
	var file *multipart.FileHeader
	if len(files) > 0 {
		file = files[0]
	}

	log.Printf("FileUpload.Process: name = %q, file = %v", c.FullName, file)
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
