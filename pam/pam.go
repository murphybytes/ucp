package pam

// #cgo LDFLAGS: -lpam
// #include <stdio.h>
// #include <stdlib.h>
// #include <security/pam_constants.h>
// int authorize_user(const char*, const char*);
import "C"

import (
	"errors"
	"unsafe"
)

var ErrUnknownUser = errors.New("Unknown user")
var ErrIncorrectPassword = errors.New("Incorrect password")
var ErrAuthFailed = errors.New("Authorization failed")

func AuthorizeUser(user, password string) error {
	cUser := C.CString(user)
	defer C.free(unsafe.Pointer(cUser))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))

	var ret C.int
	ret = C.authorize_user(cUser, cPassword)

	if ret == C.PAM_SUCCESS {
		return nil
	}

	switch ret {
	case C.PAM_USER_UNKNOWN:
		return ErrUnknownUser
	case C.PAM_AUTH_ERR:
		return ErrIncorrectPassword
	}

	return ErrAuthFailed

}
