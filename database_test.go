package recfile

import (
	"log"
	"os"
	"testing"
)

func display(d Database) {
	for _, set := range d.RecordSets {
		log.Println("Type:", set.Descriptor.Type)
		for _, property := range set.Descriptor.SpecialFields {
			log.Println("Property:", property)
		}

		log.Println()

		for _, record := range set.Records {
			for _, field := range record.Fields {
				log.Println("\tName:", field.Name, "|", "Value:", field.Value)
			}
			log.Println()
		}
	}
}

func TestLoad(harness *testing.T) {
	log.SetFlags(log.Lshortfile)

	_, err := Load("testdata/books.rec")
	if nil != err {
		harness.Log(err.Error())
		harness.Fail()
	}
}

func TestSave(harness *testing.T) {
	var err error

	d, err := Load("testdata/books.rec")
	if nil != err {
		harness.Log(err.Error())
		harness.Fail()
	}

	for i := range d.RecordSets {
		if "Magazine" == d.RecordSets[i].Descriptor.Type {
			d.RecordSets[i].Records = append(d.RecordSets[i].Records, Record{
				Fields: []Field{Field{Name: "Title", Value: "Ayyy Must Be The Money"}},
			})
			break
		}
	}

	d.RecordSets = append(d.RecordSets, RecordSet{
		Descriptor: Descriptor{
			Type: "Car",
			SpecialFields: []Property{
				Property(Field{Name: "xxx", Value: "yyy"}),
			},
		},
		Records: []Record{
			Record{Fields: []Field{Field{Name: "abc", Value: "def"}, Field{Name: "ghi", Value: "jkl"}}},
			Record{Fields: []Field{Field{Name: "mno", Value: "pqr"}, Field{Name: "stu", Value: "vwx"}}},
		},
	})

	err = d.Save("test.rec")
	if nil != err {
		harness.Log(err.Error())
		harness.Fail()
	}

	err = os.Remove("test.rec")
	if nil != err {
		harness.Log(err.Error())
		harness.Fail()
	}
}
