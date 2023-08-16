package tests

import (
	"bytes"
	"fmt"
	io2 "io"
	"os"
	"testing"

	handler "github.com/abtinokhovat/file-handler-go"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name  string
	Age   int
	Likes []int
}

var (
	testDataHandler = []testStruct{
		{
			Name:  "David",
			Age:   40,
			Likes: []int{15, 23, 35},
		},
		{
			Name:  "Eve",
			Age:   22,
			Likes: []int{12, 28, 44, 72},
		},
		{
			Name:  "Frank",
			Age:   31,
			Likes: []int{7, 19},
		},
	}
	testDataString = "Name: David\nAge: 40\nLikes: [15 23 35]\n\nName: Eve\nAge: 22\nLikes: [12 28 44 72]\n\nName: Frank\nAge: 31\nLikes: [7 19]\n\n"
)

type MockSerializer struct {
	handler.Serializer[testStruct]
}

func (s MockSerializer) Serialize(data testStruct) (string, error) {
	return fmt.Sprintf("Name: %s\nAge: %d\nLikes: %v", data.Name, data.Age, data.Likes), nil
}

func (s MockSerializer) SerializeMany(data []testStruct) (string, error) {
	var buff bytes.Buffer
	for _, value := range data {
		buff.WriteString(fmt.Sprintf("Name: %s\nAge: %d\nLikes: %v\n\n", value.Name, value.Age, value.Likes))
	}
	return buff.String(), nil
}

func (s MockSerializer) Deserialize(jsonString string) ([]testStruct, error) {
	var testData []testStruct

	data := testStruct{
		Name:  jsonString,
		Age:   404,
		Likes: nil,
	}
	testData = append(testData, data)

	return testData, nil
}

func createTempFile(data string) (*os.File, error) {
	file, err := os.CreateTemp("", "test*.json")
	if err != nil {
		return nil, err
	}

	_, err = file.WriteString(data)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func TestJsonIOHandler_File(t *testing.T) {
	t.Run("open file that does not exists", func(t *testing.T) {
		// 1. setup
		fileHandler := new(handler.JsonIOHandler[string])
		file, err := fileHandler.File("test.txt")
		defer os.Remove(file.Name())

		if err != nil {
			t.Errorf("error occured on File, %s", err)
		}

		// 2. execution
		read, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatalf("error occured on Reading the file, %s", err)
		}

		// 3. assertion
		assert.NotNil(t, read, "expected to read a file but failed")
	})

	t.Run("open file that exists", func(t *testing.T) {
		// 1. setup
		created, err := os.Create("test.txt")
		if err != nil {
			t.Errorf("error occured on creating test file, %s", err)
		}
		defer os.Remove(created.Name())

		fileHandler := new(handler.JsonIOHandler[string])
		file, err := fileHandler.File("test.txt")
		if err != nil {
			t.Fatalf("error occured on opening , %s", err)
		}

		// 2. execution
		read, err := os.ReadFile(file.Name())
		if err != nil {
			t.Errorf("error occured on Reading the file, %s", err)
		}

		// 3. assertion
		assert.NotNil(t, read, "expected to make a file but failed")
	})
}

func TestJsonIOHandler_Read(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{name: "ordinary", data: "salam"},
		{name: "empty", data: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := createTempFile(tc.data)
			if err != nil {
				return
			}
			defer file.Close()
			defer os.Remove(file.Name())

			serializer := MockSerializer{}
			fileHandler := handler.NewJsonIOHandler[testStruct](file.Name(), serializer)

			read, err := fileHandler.Read()
			if err != nil {
				t.Errorf("%v", err)
			}
			if len(read) == 0 {
				t.Errorf("List should not be empty an error may occured")
			}
			if read[0].Name != tc.data {
				t.Errorf("Expected %s, but got %s\nThe whole data:%v", read[0].Name, tc.data, read)
			}
		})
	}

}

func TestJsonIOHandler_DeleteAll(t *testing.T) {
	t.Run("ordinary", func(t *testing.T) {
		// 1. setup
		dummyText := "The data that will be deleted"
		file, err := createTempFile(dummyText)
		if err != nil {
			return
		}
		defer file.Close()
		defer os.Remove(file.Name())

		beforeDelete, err := io2.ReadAll(file)
		if err != nil || string(beforeDelete) == dummyText {
			t.Fatalf("error in the test failed to writing to file\nReadErr: %s,\n data: %v", err, beforeDelete)
		}

		serializer := MockSerializer{}
		ioHandler := handler.NewJsonIOHandler[testStruct](file.Name(), serializer)

		// 2. execution
		err = ioHandler.DeleteAll()
		if err != nil {
			t.Fatalf("Error on DeleteAll: %s", err)
		}

		// 3. assertion
		afterDelete, err := io2.ReadAll(file)
		if err != nil || len(afterDelete) != 0 {
			t.Fatalf("Falied to delete the content\nDeleteErr: %s,\n data: %v", err, afterDelete)
		}
	})

	t.Run("not_valid_file", func(t *testing.T) {
		// 1. setup
		serializer := MockSerializer{}
		filePath := "non_existent_file.json"
		ioHandler := handler.NewJsonIOHandler[testStruct](filePath, serializer)

		// 2. execution
		err := ioHandler.DeleteAll()
		if err != nil {
			t.Errorf("failed on deleting the content in the file %s", err)
		}

		file, err := os.OpenFile(filePath, os.O_RDWR, 0777)
		if err != nil {
			t.Errorf("failed on openning the file %s", err)
		}
		defer file.Close()
		defer os.Remove(filePath)

		content, err := io2.ReadAll(file)
		if err != nil {
			t.Errorf("failed on reading the content in the file %s", err)
		}

		// 3. assertion
		assert.Equalf(t, len(content), 0, "Falied to delete the content\nDeleteErr: %s,\n data: %v", err, content)

	})
}

func TestJsonIOHandler_DeleteAndWrite(t *testing.T) {
	t.Run("ordinary", func(t *testing.T) {
		// 1. setup
		initText := "The init data is here and here"

		file, err := createTempFile(initText)
		if err != nil {
			return
		}
		defer file.Close()
		defer os.Remove(file.Name())

		beforeDeleteAndWrite, err := io2.ReadAll(file)
		if err != nil || string(beforeDeleteAndWrite) == initText {
			t.Fatalf("error in the test failed to writing to file\nReadErr: %s,\n data: %v", err, beforeDeleteAndWrite)
		}

		serializer := MockSerializer{}
		ioHandler := handler.NewJsonIOHandler[testStruct](file.Name(), serializer)

		// 2. execution
		err = ioHandler.DeleteAndWrite(testDataHandler)
		if err != nil {
			t.Fatalf("Error on DeleteAndWrite: %s", err)
		}
		_, _ = file.Seek(0, 0)
		afterDelete, err := io2.ReadAll(file)

		// 3. assertion
		assert.Equal(t, testDataString, string(afterDelete), fmt.Sprintf("Falied to delete and rewrite the content\nDeleteAndWrite: %s,\n data: %v", err, afterDelete))
	})

	t.Run("not_valid_file", func(t *testing.T) {
		// 1. setup
		serializer := MockSerializer{}
		filePath := "non_existent_file.json"
		ioHandler := handler.NewJsonIOHandler[testStruct](filePath, serializer)

		// 2. execution
		err := ioHandler.DeleteAndWrite(testDataHandler)
		if err != nil {
			t.Errorf("failed on deleting the content in the file %s", err)
		}

		file, err := os.OpenFile(filePath, os.O_RDWR, 0777)
		if err != nil {
			t.Errorf("failed on openning the file %s", err)
		}
		defer file.Close()
		defer os.Remove(filePath)

		content, err := io2.ReadAll(file)
		if err != nil {
			t.Errorf("failed on reading the content in the file %s", err)
		}

		// 3. assertion
		assert.Equalf(t, testDataString, string(content), "Falied to delete the content\nDeleteErr: %s,\n data: %v", err, content)
	})
}
