# sqlite-go-model-generator

I'm dealing with a lot of sqlite databases. This code base help me to generate go structures with orm tags on the structre. 

Though I can use this [JSON-to-GO](https://mholt.github.io/json-to-go/) tool to generate structures to represent the databbase tables , there are two manually steps I have to do:

1. convert one of the table row data to json by something like [CSV to JSON](https://csvjson.com/), to generate a json out of one of the database row
2. Paste the generated JSON string to [JSON-to-GO](https://mholt.github.io/json-to-go/), and copy the generated go structure code. I will get something like this 
  ```go
    type AutoGenerated struct {
  	  A string `json:"a"`
  	  D string `json:"d"`
    }
  ```
3. Add tags to each structure fields with something like ```gorm:"column:data"```. Then I can do db query/updating by simply using [ORM](https://github.com/jinzhu/gorm).

When there are tens of columns, adding orm tag is still a lot of error prone manually work. 

This tool querys the ```sqlite_master``` database for all tables, and query ```table_info``` for the columns and their types. Then [Jenifer](https://github.com/dave/jennifer) helps to do codegen, and save the generated code under ```gen``` folder.

```go
package def

// Test represent database table (test)
type Test struct {
	ID   int32 `gorm:"column:id" json:"id"`
	Data int32 `gorm:"column:data" json:"data"`
}

// TableName represent the database table name of Test
func (Test) TableName() string {
	return "test"
}
```

# Steps to run

Clone the code to local and do the build.

```sh
go build
./sqlite-go-model-generator
```
