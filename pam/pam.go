package pam

// #cgo LDFLAGS: -lpam
// #include <stdio.h>
// #include <stdlib.h>
// #include <security/pam_constants.h>
// int authorize_user(const char*, const char*);
import "C"

import (
	"fmt"
	"unsafe"
)

func AuthorizeUser(user, password string) (auth bool, e error) {
	cUser := C.CString(user)
	defer C.free(unsafe.Pointer(cUser))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))

	var ret C.int
	ret = C.authorize_user(cUser, cPassword)

	auth = (ret == C.PAM_SUCCESS)

	if ret != C.PAM_SUCCESS && ret != C.PAM_AUTH_ERR {
		e = fmt.Errorf("Authorization error %d\n", ret)
	}

	return

}
