package main

import (
	"encoding/xml"
	"fmt"
	"tank/rest"
	"testing"
)

func TestXmlMarshal(t *testing.T) {
	fmt.Println("Hello World!")

	type Address struct {
		City, State string
	}
	type Person struct {
		XMLName   xml.Name `xml:"person"`
		Id        int      `xml:"id,attr"`
		FirstName string   `xml:"name>first"`
		LastName  string   `xml:"name>last"`
		Age       int      `xml:"age"`
		Height    float32  `xml:"height,omitempty"`
		Married   bool
		Address
		Comment string `xml:",comment"`
	}

	v := &Person{Id: 13, FirstName: "John", LastName: "Doe", Age: 42}
	v.Comment = " Need more details. "
	v.Address = Address{"Hanga Roa", "Easter Island"}

	output, err := xml.MarshalIndent(v, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Println(string(output))
	// Output:
	//   <person id="13">
	//       <name>
	//           <first>John</first>
	//           <last>Doe</last>
	//       </name>
	//       <age>42</age>
	//       <Married>false</Married>
	//       <City>Hanga Roa</City>
	//       <State>Easter Island</State>
	//       <!-- Need more details. -->
	//   </person>

}

func TestXmlUnMarshal(t *testing.T) {
	type Email struct {
		Where string `xml:"where,attr"`
		Addr  string
	}
	type Address struct {
		City, State string
	}
	type Result struct {
		XMLName xml.Name `xml:"Person"`
		Name    string   `xml:"FullName"`
		Phone   string
		Email   []Email
		Groups  []string `xml:"Group>Value"`
		Address
	}
	v := Result{Name: "none", Phone: "none"}

	data := `
		<Person>
			<FullName>Grace R. Emlin</FullName>
			<Company>Example Inc.</Company>
			<Email where="home">
				<Addr>gre@example.com</Addr>
			</Email>
			<Email where='work'>
				<Addr>gre@work.com</Addr>
			</Email>
			<Group>
				<Value>Friends</Value>
				<Value>Squash</Value>
			</Group>
			<City>Hanga Roa</City>
			<State>Easter Island</State>
		</Person>
	`
	err := xml.Unmarshal([]byte(data), &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Printf("XMLName: %#v\n", v.XMLName)
	fmt.Printf("Name: %q\n", v.Name)
	fmt.Printf("Phone: %q\n", v.Phone)
	fmt.Printf("Email: %v\n", v.Email)
	fmt.Printf("Groups: %v\n", v.Groups)
	fmt.Printf("Address: %v\n", v.Address)
	// Output:
	// XMLName: xml.Name{Space:"", Local:"Person"}
	// Name: "Grace R. Emlin"
	// Phone: "none"
	// Email: [{home gre@example.com} {work gre@work.com}]
	// Groups: [Friends Squash]
	// Address: {Hanga Roa Easter Island}
}

func TestPropfind(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8" ?>
		<D:propfind xmlns:D="DAV:">
			<D:prop>
				<D:resourcetype />
				<D:getcontentlength />
				<D:creationdate />
				<D:getlastmodified />
			</D:prop>
		</D:propfind>
	`
	propfind := &rest.Propfind{XmlNS: "DAV:"}
	err := xml.Unmarshal([]byte(data), &propfind)
	if err != nil {
		panic(err)
	}

	output, err := xml.MarshalIndent(propfind, "  ", "    ")
	fmt.Println(string(output))
}
