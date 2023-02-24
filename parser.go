package recfile

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	LineDelimiterRune     = '\n'
	CommentPrefix         = "#"
	FieldSeparator        = ": "
	LineWrapPrefix        = "+"
	DefaultRecordType     = ""
	SpecialFieldPrefix    = "%"
	ValidFieldNamePattern = `[a-zA-Z%][a-zA-Z0-9_]*` // regular expression provided by GNU recutils documentation
	ValidTypeNamePattern  = `[a-zA-Z][a-zA-Z0-9_]*`  // regular expression provided by GNU recutils documentation

	RecProperty          = "rec"
	MandatoryProperty    = "mandatory"
	AllowedProperty      = "allowed"
	ProhibitProperty     = "prohibit"
	UniqueProperty       = "unique"
	KeyProperty          = "key"
	DocProperty          = "doc"
	TypedefProperty      = "typedef"
	TypeProperty         = "type"
	AutoProperty         = "auto"
	SortProperty         = "sort"
	SizeProperty         = "size"
	ConstraintProperty   = "constraint"
	ConfidentialProperty = "confidential"

	LessThan             = "<"
	LessThanOrEqualTo    = "<="
	GreaterThan          = ">"
	GreaterThanOrEqualTo = ">="
)

type parser struct {
	*bufio.Reader
	file                *os.File
	properties          map[string]string
	relationalOperators map[string]string
}

func newParser(path string) (*parser, error) {
	var reader parser
	var err error

	file, err := os.Open(path)
	if nil != err {
		return &reader, err
	}

	reader.Reader = bufio.NewReader(file)
	reader.file = file
	reader.properties = map[string]string{
		RecProperty:          RecProperty,
		MandatoryProperty:    MandatoryProperty,
		AllowedProperty:      AllowedProperty,
		ProhibitProperty:     ProhibitProperty,
		UniqueProperty:       UniqueProperty,
		KeyProperty:          KeyProperty,
		DocProperty:          DocProperty,
		TypedefProperty:      TypedefProperty,
		TypeProperty:         TypeProperty,
		AutoProperty:         AutoProperty,
		SortProperty:         SortProperty,
		SizeProperty:         SizeProperty,
		ConstraintProperty:   ConstraintProperty,
		ConfidentialProperty: ConfidentialProperty,
	}
	reader.relationalOperators = map[string]string{
		LessThan:             LessThan,
		LessThanOrEqualTo:    LessThanOrEqualTo,
		GreaterThan:          GreaterThan,
		GreaterThanOrEqualTo: GreaterThanOrEqualTo,
	}

	return &reader, err
}

func (reader *parser) closeParser() error {
	var err error

	err = reader.file.Close()
	if nil != err {
		return err
	}

	return err
}

/* Returns a logical recfile line */
func (reader *parser) getLine() (string, error) {
	var line string
	var err error

	/* After this loop terminates, line will be the first non-comment line */
	for line, err = reader.ReadString(LineDelimiterRune); nil == err && true == strings.HasPrefix(line, CommentPrefix); line, err = reader.ReadString(LineDelimiterRune) {
	}
	line = strings.Replace(line, "\n", "", 1)

	/* Now we peek to see if line wraps to the next line of the file */
	var peeked []byte
	for peeked, err = reader.Peek(len(LineWrapPrefix)); nil == err; peeked, err = reader.Peek(len(LineWrapPrefix)) {
		if true == strings.HasPrefix(string(peeked), CommentPrefix) {
			_, err = reader.ReadString(LineDelimiterRune) // discard comments
			if nil != err {
				break
			}
			continue
		}

		if true == strings.HasPrefix(string(peeked), LineWrapPrefix) {
			if "" == line {
				err = errors.New("Found invalid line wrap marker")
				break
			}

			nextline, err := reader.ReadString(LineDelimiterRune)
			if nil != err {
				break
			}

			line += strings.Replace(strings.Replace(nextline, LineWrapPrefix, "", 1), "\n", "", 1)
		} else {
			break
		}
	}

	return line, err
}

func (reader *parser) getRecord() (Record, error) {
	var record Record
	var err error

	var line string
	for line, err = reader.getLine(); (nil == err || io.EOF == err) && "" != line; line, err = reader.getLine() {
		split := strings.Split(line, FieldSeparator)
		if 2 != len(split) {
			err = errors.New("Found invalid field")
			break
		}

		var matched bool
		matched, err = regexp.MatchString(ValidFieldNamePattern, split[0])
		if false == matched {
			if nil == err {
				err = errors.New("Found invalid field name")
			}
			break
		}

		record.Fields = append(record.Fields, Field{Name: Name(split[0]), Value: Value(split[1])})
	}

	return record, err
}

func (reader *parser) validateProperty(property Property) error {
	var err error

	split := strings.Fields(string(property.Value))

	switch property.Name {
	case SpecialFieldPrefix + RecProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}

		if 2 == len(split) {
			_, err = url.Parse(split[1])
			break
		}
	case SpecialFieldPrefix + MandatoryProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + AllowedProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + ProhibitProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + UniqueProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + KeyProperty:
		if 1 != len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + DocProperty:
	case SpecialFieldPrefix + TypedefProperty:
		if 2 > len(split) {
			err = errors.New("Missing type name and/or type description")
			break
		}

		var matched bool
		matched, err = regexp.MatchString(ValidTypeNamePattern, split[0])
		if false == matched {
			if nil == err {
				err = errors.New("Found invalid field name")
			}
			break
		}
	case SpecialFieldPrefix + TypeProperty: // to do
		if 2 > len(split) {
			err = errors.New("Missing field list, type name, or type description")
			break
		}
	case SpecialFieldPrefix + AutoProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + SortProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + SizeProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}

		if 2 < len(split) {
			err = errors.New("Too many arguments found")
			break
		}

		_, err = strconv.Atoi(split[len(split)-1]) // we don't care what number it is
		if nil != err {
			break
		}

		if 2 == len(split) {
			var present bool
			_, present = reader.relationalOperators[split[0]]
			if true == present {
				err = errors.New("Found invalid relational operator")
				break
			}
		}
	case SpecialFieldPrefix + ConstraintProperty:
		if 0 == len(split) {
			err = errors.New("No property value found")
			break
		}
	case SpecialFieldPrefix + ConfidentialProperty:
	}

	return err
}

func (reader *parser) getRecordSetDescriptor() (Descriptor, error) {
	var record Descriptor
	var err error

	var line string
	var flag bool
	for line, err = reader.getLine(); (nil == err || io.EOF == err) && "" != line; line, err = reader.getLine() {
		split := strings.Split(line, FieldSeparator)
		if 2 != len(split) {
			err = errors.New("Found invalid property")
			break
		}

		property := Property(Field{Name: Name(strings.TrimPrefix(split[0], SpecialFieldPrefix)), Value: Value(split[1])})

		_, present := reader.properties[string(property.Name)]
		if true == present {
			err = reader.validateProperty(property)
			if nil != err {
				break
			}
		}

		if RecProperty == property.Name {
			record.Type = Type(property.Value)
			if true == flag {
				err = errors.New("Found multiple record types")
				break
			}
			flag = true
		} else {
			record.SpecialFields = append(record.SpecialFields, property)
		}
	}

	if false == flag {
		err = errors.New("Missing record type")
	}

	return record, err
}

func (reader *parser) getRecordSet() (RecordSet, error) {
	var recordSet RecordSet
	var err error

	recordSet.Descriptor.Type = Type(DefaultRecordType)
	for {
		var peeked []byte
		peeked, err = reader.Peek(len(SpecialFieldPrefix))
		if nil != err {
			break
		}

		if CommentPrefix == string(peeked) {
			_, err = reader.ReadString(LineDelimiterRune) // discard comments
			if nil != err {
				break
			}
			continue
		} else if SpecialFieldPrefix == string(peeked) {
			/* If we already have records in our record set or we
			already have a non-default record set descriptor, then
			encountering a SpecialFieldPrefix character means the
			upcoming record must be a record set descriptor for a
			new record set. Thus, we end our parsing */
			if 0 != len(recordSet.Records) || Type(DefaultRecordType) != recordSet.Descriptor.Type {
				break
			}

			var descriptor Descriptor
			descriptor, err = reader.getRecordSetDescriptor()
			if nil != err && io.EOF != err {
				break
			}
			recordSet.Descriptor = descriptor
		} else if SpecialFieldPrefix != string(peeked) {
			var record Record
			record, err = reader.getRecord()
			if 0 < len(record.Fields) {
				recordSet.Records = append(recordSet.Records, record)
			}

			if nil != err {
				break
			}
		}
	}

	return recordSet, err
}

func (reader *parser) getDatabase() (Database, error) {
	var db Database
	var err error

	for {
		var recordSet RecordSet
		recordSet, err = reader.getRecordSet()
		if 0 < len(recordSet.Records) || Type(DefaultRecordType) != recordSet.Descriptor.Type {
			db.RecordSets = append(db.RecordSets, recordSet)
		}

		if nil != err {
			break
		}
	}

	/* End of File is not a report-worthy error, as it actually
	indicates that the entire recfile was parsed successfully */
	if io.EOF == err {
		err = nil
	}

	return db, err
}
