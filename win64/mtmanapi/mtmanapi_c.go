package mtmanapi

// #cgo windows LDFLAGS: -lws2_32 -lstdc++ -static
// #cgo windows,amd64 CXXFLAGS: -target x86_64-pc-windows-msvc
import "C"
