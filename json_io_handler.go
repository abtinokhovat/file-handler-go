package handler

import (
	"io"
	"os"
)

type FileReader[T any] interface {
	Read() ([]T, error)
}

type FileWriter[T any] interface {
	WriteOne(data T) error
	DeleteAll() error
	DeleteAndWrite(data []T) error
}

type FileIOHandler[T any] interface {
	FileReader[T]
	FileWriter[T]
}

type JsonIOHandler[T any] struct {
	FileIOHandler[T]
	filePath   string
	serializer Serializer[T]
}

func NewJsonIOHandler[T any](path string, serializer Serializer[T]) *JsonIOHandler[T] {
	return &JsonIOHandler[T]{
		filePath:   path,
		serializer: serializer,
	}
}

func (h *JsonIOHandler[T]) File(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (h *JsonIOHandler[T]) openFile() (*os.File, error) {
	return h.File(h.filePath)
}

// Should read a file and return the data in file with type of data
func (h *JsonIOHandler[T]) Read() ([]T, error) {
	file, err := h.openFile()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	deserialized, err := h.serializer.Deserialize(string(content))
	if err != nil {
		return nil, err
	}

	return deserialized, nil
}

func (h *JsonIOHandler[T]) WriteOne(data T) error {
	// reading data
	content, err := h.Read()
	if err != nil {
		return err
	}

	content = append(content, data)

	err = h.DeleteAndWrite(content)
	if err != nil {
		return err
	}

	return nil
}

func (h *JsonIOHandler[T]) DeleteAndWrite(data []T) error {
	// open file for writing
	file, err := h.openFile()
	if err != nil {
		return err
	}
	defer file.Close()

	// remove data
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	str, err := h.serializer.SerializeMany(data)
	if err != nil {
		return err
	}
	_, err = file.WriteString(str)
	if err != nil {
		return err
	}

	return nil
}

func (h *JsonIOHandler[T]) DeleteAll() error {
	file, err := h.openFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return err
	}

	return nil
}
