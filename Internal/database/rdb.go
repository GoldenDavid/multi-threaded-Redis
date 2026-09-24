package database

import (
	"container/list"
	"encoding/json"
	"os"
)

// rdbEntity represents the JSON structure for RDB snapshotting.
type rdbEntity struct {
	Type DataType    `json:"type"`
	Val  interface{} `json:"val"`
}

// SaveRDB serializes the current database state to a JSON file.
func (db *Database) SaveRDB(filename string) error {
	snapshot := make(map[string]rdbEntity)

	for key, entity := range db.data {
		var val interface{}
		switch entity.Type {
		case TypeString:
			val = entity.Val.(string)
		case TypeHash:
			val = entity.Val.(map[string]string)
		case TypeList:
			l := entity.Val.(*list.List)
			var elements []string
			for e := l.Front(); e != nil; e = e.Next() {
				elements = append(elements, e.Value.(string))
			}
			val = elements
		}
		snapshot[key] = rdbEntity{
			Type: entity.Type,
			Val:  val,
		}
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadRDB deserializes the database state from a JSON file.
func (db *Database) LoadRDB(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var snapshot map[string]rdbEntity
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	db.data = make(map[string]DataEntity)

	for key, entity := range snapshot {
		var val interface{}
		switch entity.Type {
		case TypeString:
			val = entity.Val.(string)
		case TypeHash:
			// JSON unmarshals maps into map[string]interface{}, convert to map[string]string
			m := entity.Val.(map[string]interface{})
			hash := make(map[string]string)
			for k, v := range m {
				hash[k] = v.(string)
			}
			val = hash
		case TypeList:
			// JSON unmarshals arrays into []interface{}
			arr := entity.Val.([]interface{})
			l := list.New()
			for _, item := range arr {
				l.PushBack(item.(string))
			}
			val = l
		}
		db.data[key] = DataEntity{
			Type: entity.Type,
			Val:  val,
		}
	}

	return nil
}
