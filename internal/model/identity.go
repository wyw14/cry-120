package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Identity string

func NewIdentity(prefix string) Identity {
	return Identity(fmt.Sprintf("%s-%s", prefix, uuid.NewString()))
}

func (i Identity) String() string {
	return string(i)
}

func (i Identity) Empty() bool {
	return i == ""
}

type OperationID Identity

func NewOperationID() OperationID {
	return OperationID(NewIdentity("op"))
}

func (i OperationID) String() string {
	return string(i)
}

type Revision Identity

func NewRevision() Revision {
	return Revision(NewIdentity("rev"))
}

func (r Revision) String() string {
	return string(r)
}

type Token Identity

func NewToken(prefix string) Token {
	return Token(NewIdentity(prefix))
}

func (t Token) String() string {
	return string(t)
}
