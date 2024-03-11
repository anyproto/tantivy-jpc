package tantivy

// #cgo windows,amd64 LDFLAGS:-L${SRCDIR}/packaged/lib/windows-amd64 -ltantivy_jpc -lm -pthread -lws2_32 -lbcrypt -lwsock32 -lntdll -luserenv -lsynchronization
// #cgo darwin,amd64 LDFLAGS:-L${SRCDIR}/packaged/lib/darwin-amd64 -ltantivy_jpc -lm -pthread -framework CoreFoundation -framework Security -ldl
// #cgo darwin,arm64 LDFLAGS:-L${SRCDIR}/packaged/lib/darwin-aarch64 -ltantivy_jpc -lm -pthread -ldl
// #cgo ios,arm64 LDFLAGS:-L${SRCDIR}/packaged/lib/ios-aarch64 -ltantivy_jpc -lm -pthread -ldl
// #cgo ios,amd64 LDFLAGS:-L${SRCDIR}/packaged/lib/ios-amd64 -ltantivy_jpc -lm -pthread -ldl
// #cgo android,arm LDFLAGS:-L${SRCDIR}/packaged/lib/android-arm -ltantivy_jpc -lm -ldl
// #cgo android,386 LDFLAGS:-L${SRCDIR}/packaged/lib/android-x86 -ltantivy_jpc -lm -ldl
// #cgo android,amd64 LDFLAGS:-L${SRCDIR}/packaged/lib/android-amd64 -ltantivy_jpc -lm -ldl
// #cgo android,arm64 LDFLAGS:-L${SRCDIR}/packaged/lib/android-arm64 -ltantivy_jpc -lm -ldl
// #cgo CFLAGS: -I${SRCDIR}/packaged/include
// #cgo linux,amd64,!android LDFLAGS:-L${SRCDIR}/packaged/lib/linux-amd64 -Wl,--allow-multiple-definition -ltantivy_jpc -lm -pthread -lpthread
//
// #include "tantivy-jpc.h"
// #include <stdlib.h>
import "C"
import (
	"encoding/json"
	"os"
	"sync"
	"unsafe"

	"github.com/eluv-io/errors-go"
)

var doOnce sync.Once

func LibInit(directive ...string) {
	var initVal string
	doOnce.Do(func() {
		if len(directive) == 0 {
			initVal = "info"
		} else {
			initVal = directive[0]
		}
		os.Setenv("ELV_RUST_LOG", initVal)
		C.init()
	})
}

func ClearSession(sessionID string) {
	C.term(C.CString(sessionID))
}

func SetKB(k float64, b float64) {
	C.set_k_and_b(C.float(k), C.float(b))
}

type msi = map[string]interface{}

const defaultMemSize = uint32(500000000)

// The ccomsBuf is a raw byte buffer for tantivy-jpc to send results. A single mutex guards its use.
type JPCId struct {
	id       string
	TempDir  string
	ccomsBuf *C.char
	bufLen   int32
}

func (j *JPCId) ID() string {
	return j.id
}

func (jpc *JPCId) callTantivy(object, method string, params msi) (string, error) {
	f := map[string]interface{}{
		"id":     jpc.id,
		"jpc":    "1.0",
		"obj":    object,
		"method": method,
		"params": params,
	}
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	var pcomsBuf *C.char
	var blen int64
	sb := string(b)
	pcJPCParams := C.CString(sb)
	pCDesctination := (*C.uchar)(unsafe.Pointer(pcomsBuf))
	cJPCParams := (*C.uchar)(unsafe.Pointer(pcJPCParams))
	pDestinationLen := (*pointerCType)(unsafe.Pointer(&blen))
	ttret := C.tantivy_jpc(cJPCParams, pointerCType(pointerGoType(len(sb))), &pCDesctination, pDestinationLen)
	if ttret < 0 {
		return "", errors.E("Tantivy JPC Failed", errors.K.Invalid, "desc", string(C.GoBytes(unsafe.Pointer(pCDesctination), C.int(*pDestinationLen))))
	}
	defer func() {
		if ttret >= 0 {
			C.free_data(ttret)
		}
	}()
	returnData := string(C.GoBytes(unsafe.Pointer(pCDesctination), C.int(*pDestinationLen)))
	return returnData, nil
}
