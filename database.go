package recfile

import (
	"errors"
	"os"
)

type Name string
type Value string
type Type string

type Field struct {
	Name  Name
	Value Value
}

type Property Field

type Descriptor struct {
	Type          Type
	SpecialFields []Property
}

type Record struct {
	Fields []Field
}

type RecordSet struct {
	Descriptor Descriptor
	Records    []Record
}

type Database struct {
	RecordSets []RecordSet
}

func Load(path string) (Database, error) {
	var database Database
	var err error

	parser, err := newParser(path)
	if nil != err {
		return database, err
	}
	defer parser.closeParser()

	database, err = parser.getDatabase()
	if nil != err {
		return database, err
	}

	return database, err
}

func (database *Database) Save(path string) error {
	var err error

	target, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if nil != err {
		return err
	}
	defer target.Close()

	for _, set := range database.RecordSets {
		var flag bool

		/* From GNU recutils documentation:
		Every record set must contain one, and only one, record set descriptor type */
		if DefaultRecordType == set.Descriptor.Type && 0 != len(set.Descriptor.SpecialFields) {
			err = errors.New("Invalid record set descriptor")
			return err
		}

		if DefaultRecordType != set.Descriptor.Type {
			_, err = target.WriteString(SpecialFieldPrefix + RecProperty + FieldSeparator + string(set.Descriptor.Type) + string(LineDelimiterRune))
			if nil != err {
				return err
			}

			flag = true

			for _, property := range set.Descriptor.SpecialFields {
				_, err = target.WriteString(SpecialFieldPrefix + string(property.Name) + FieldSeparator + string(property.Value) + string(LineDelimiterRune))
				if nil != err {
					return err
				}

				flag = true
			}
		}

		if true == flag { // don't need a newline if there was no descriptor
			_, err = target.WriteString(string(LineDelimiterRune))
			if nil != err {
				return err
			}

			flag = false
		}

		for i, record := range set.Records {
			/* Add an index check to determine if an extra newline needs to be written */
			for _, field := range record.Fields {
				_, err = target.WriteString(string(field.Name) + FieldSeparator + string(field.Value) + string(LineDelimiterRune))
				if nil != err {
					return err
				}
			}

			if (len(set.Records) - 1) != i {
				_, err = target.WriteString(string(LineDelimiterRune))
				if nil != err {
					return err
				}
			}
		}
	}

	return err
}
