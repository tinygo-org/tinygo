package macos

import (
	"errors"
	"time"
)

// Exported symbols copied from Big Go, but stripped of functionality.
// Allows building of crypto/x509 on macOS.

const (
	ErrSecCertificateExpired = -67818
	ErrSecHostNameMismatch   = -67602
	ErrSecNotTrusted         = -67843
)

var ErrNoTrustSettings = errors.New("no trust settings found")
var SecPolicyAppleSSL = StringToCFString("1.2.840.113635.100.1.3") // defined by POLICYMACRO
var SecPolicyOid = StringToCFString("SecPolicyOid")
var SecTrustSettingsPolicy = StringToCFString("kSecTrustSettingsPolicy")
var SecTrustSettingsPolicyString = StringToCFString("kSecTrustSettingsPolicyString")
var SecTrustSettingsResultKey = StringToCFString("kSecTrustSettingsResult")

func CFArrayAppendValue(array CFRef, val CFRef) {}

func CFArrayGetCount(array CFRef) int {
	return 0
}

func CFDataGetBytePtr(data CFRef) uintptr {
	return 0
}

func CFDataGetLength(data CFRef) int {
	return 0
}

func CFDataToSlice(data CFRef) []byte {
	return nil
}

func CFEqual(a, b CFRef) bool {
	return false
}

func CFErrorGetCode(errRef CFRef) int {
	return 0
}

func CFNumberGetValue(num CFRef) (int32, error) {
	return 0, errors.New("not implemented")
}

func CFRelease(ref CFRef) {}

func CFStringToString(ref CFRef) string {
	return ""
}

func ReleaseCFArray(array CFRef) {}

func SecCertificateCopyData(cert CFRef) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func SecTrustCopyCertificateChain(trustObj CFRef) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecTrustEvaluateWithError(trustObj CFRef) (int, error) {
	return 0, errors.New("not implemented")
}

func SecTrustGetCertificateCount(trustObj CFRef) int {
	return 0
}

func SecTrustGetResult(trustObj CFRef, result CFRef) (CFRef, CFRef, error) {
	return 0, 0, errors.New("not implemented")
}

func SecTrustSetVerifyDate(trustObj CFRef, dateRef CFRef) error {
	return errors.New("not implemented")
}

type CFRef uintptr

func BytesToCFData(b []byte) CFRef {
	return 0
}

func CFArrayCreateMutable() CFRef {
	return 0
}

func CFArrayGetValueAtIndex(array CFRef, index int) CFRef {
	return 0
}

func CFDateCreate(seconds float64) CFRef {
	return 0
}

func CFDictionaryGetValueIfPresent(dict CFRef, key CFString) (value CFRef, ok bool) {
	return 0, false
}

func CFErrorCopyDescription(errRef CFRef) CFRef {
	return 0
}

func CFStringCreateExternalRepresentation(strRef CFRef) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecCertificateCreateWithData(b []byte) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecPolicyCreateSSL(name string) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecTrustCreateWithCertificates(certs CFRef, policies CFRef) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecTrustEvaluate(trustObj CFRef) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecTrustGetCertificateAtIndex(trustObj CFRef, i int) (CFRef, error) {
	return 0, errors.New("not implemented")
}

func SecTrustSettingsCopyCertificates(domain SecTrustSettingsDomain) (certArray CFRef, err error) {
	return 0, errors.New("not implemented")
}

func SecTrustSettingsCopyTrustSettings(cert CFRef, domain SecTrustSettingsDomain) (trustSettings CFRef, err error) {
	return 0, errors.New("not implemented")
}

func TimeToCFDateRef(t time.Time) CFRef {
	return 0
}

type CFString CFRef

func StringToCFString(s string) CFString {
	return 0
}

type OSStatus struct {
	// Has unexported fields.
}

func (s OSStatus) Error() string

type SecTrustResultType int32

const (
	SecTrustResultInvalid SecTrustResultType = iota
	SecTrustResultProceed
	SecTrustResultConfirm // deprecated
	SecTrustResultDeny
	SecTrustResultUnspecified
	SecTrustResultRecoverableTrustFailure
	SecTrustResultFatalTrustFailure
	SecTrustResultOtherError
)

type SecTrustSettingsDomain int32

const (
	SecTrustSettingsDomainUser SecTrustSettingsDomain = iota
	SecTrustSettingsDomainAdmin
	SecTrustSettingsDomainSystem
)

type SecTrustSettingsResult int32

const (
	SecTrustSettingsResultInvalid SecTrustSettingsResult = iota
	SecTrustSettingsResultTrustRoot
	SecTrustSettingsResultTrustAsRoot
	SecTrustSettingsResultDeny
	SecTrustSettingsResultUnspecified
)
