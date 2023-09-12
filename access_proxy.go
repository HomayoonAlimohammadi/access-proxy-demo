package main

import (
	"fmt"
	"log"
)

type User struct {
	Name *string
	ID   int
}

type AccessProxy struct {
	original      User
	data          *User
	changedFields map[string]struct{}
	changeSigs    []changeSig
}

func NewAccessProxy(data *User) *AccessProxy {
	return &AccessProxy{
		data:          data,
		original:      *data,
		changedFields: make(map[string]struct{}),
	}
}

type changeSig interface {
	fieldName() string
}

type nameChangeSig struct {
	val *string
}

func (s nameChangeSig) fieldName() string { return "Name" }

type idChangeSig struct {
	val int
}

func (s idChangeSig) fieldName() string { return "ID" }

func (p *AccessProxy) GetName() *string {
	copy := *p.data.Name
	return &copy
}

func (p *AccessProxy) SetName(val *string) {
	p.changeSigs = append(p.changeSigs, nameChangeSig{val: val})
}

func (p *AccessProxy) GetID() int {
	return p.data.ID
}

func (p *AccessProxy) SetID(val int) {
	p.changeSigs = append(p.changeSigs, idChangeSig{val: val})
}

func (p *AccessProxy) rollback() {
	*p.data = p.original
}

func (p *AccessProxy) ApplyChanges() error {
	for _, s := range p.changeSigs {

		if _, alreadyChanged := p.changedFields[s.fieldName()]; alreadyChanged {
			p.rollback()
			return fmt.Errorf("duplicate changes detected for field `%s`", s.fieldName())
		}

		p.changedFields[s.fieldName()] = struct{}{}

		switch sig := s.(type) {
		case nameChangeSig:
			p.data.Name = sig.val
		case idChangeSig:
			p.data.ID = sig.val
		default:
			// should not happen
			return fmt.Errorf("unknown signal: %+v", sig)
		}
	}

	return nil
}

func main() {
	usr := &User{Name: toPointer("name"), ID: 1}
	ap := NewAccessProxy(usr)

	ap.SetName(toPointer("changed name"))
	ap.SetID(2)
	// ap.SetName(toPointer("changed 2 name")) // un comment to cause an error

	err := ap.ApplyChanges()
	if err != nil {
		fmt.Printf("failed: usr: %+v\n", usr)
		fmt.Println("usr.Name:", *usr.Name)
		log.Fatal(err)
	}

	fmt.Printf("usr: %+v\n", usr)
	fmt.Println("usr.Name:", *usr.Name)

	n := ap.GetName()
	*n = "BAD CHANGE"
	fmt.Printf("%s\n", *usr.Name) // should not be changed to BAD CHANGE
}

func toPointer[T any](t T) *T {
	return &t
}
