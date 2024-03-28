package resolver

import "errors"

// ErrNilHTTPClientWrapper signals that a nil http client wrapper was provided
var ErrNilHTTPClientWrapper = errors.New("nil http client wrapper")

// ErrNilKeysGenerator signals that a nil keys generator was provided
var ErrNilKeysGenerator = errors.New("nil keys generator")

// ErrNilPubKeyConverter signals that a nil pub key converter was provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilMarshaller signals that a nil marshaller was provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilHasher signals that a nil hasher was provided
var ErrNilHasher = errors.New("nil hasher")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidGuardian signals that the given guardian is not valid
var ErrInvalidGuardian = errors.New("invalid guardian")

// ErrInvalidSender signals that the given sender is not valid
var ErrInvalidSender = errors.New("invalid sender")

// ErrGuardianNotUsable signals that the given guardian is not yet usable
var ErrGuardianNotUsable = errors.New("guardian not yet usable")

// ErrGuardianMismatch signals that a guardian mismatch was detected on transactions
var ErrGuardianMismatch = errors.New("guardian mismatch")

// ErrNilSignatureVerifier signals that a nil signature verifier was provided
var ErrNilSignatureVerifier = errors.New("nil signature verifier")

// ErrNilGuardedTxBuilder signals that a nil guarded tx builder was provided
var ErrNilGuardedTxBuilder = errors.New("nil guarded tx builder")

// ErrNilDB signals that a nil db was provided
var ErrNilDB = errors.New("nil db")

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrNilCryptoComponentsHolderFactory signals that a nil crypto components holder factory was provided
var ErrNilCryptoComponentsHolderFactory = errors.New("nil crypto components holder factory")

// ErrNoBalance signals that the provided account has no balance
var ErrNoBalance = errors.New("no balance")

// ErrNilTOTPHandler signals that a nil TOTP handler was provided
var ErrNilTOTPHandler = errors.New("nil TOTP handler")

// ErrNilSecureOtpHandler signals that a nil secure TOTP handler was provided
var ErrNilSecureOtpHandler = errors.New("nil secure TOTP handler")

// ErrNilUserInfo signals that a nil user info was provided
var ErrNilUserInfo = errors.New("nil user info")

// ErrNilEncryptor signals that a nil encryptor was provided
var ErrNilEncryptor = errors.New("nil encryptor")

// ErrNilUserEncryptor signals that a nil user encryptor was provided
var ErrNilUserEncryptor = errors.New("nil user encryptor")

// ErrTooManyTransactionsToSign signals that too many transactions were provided to sign
var ErrTooManyTransactionsToSign = errors.New("too many transactions to sign")

// ErrNoTransactionToSign signals that no transaction was provided to sign
var ErrNoTransactionToSign = errors.New("no transaction to sign")

// ErrEmptyUrl signals that an empty url was provided
var ErrEmptyUrl = errors.New("empty url")

// ErrSecondCodeInvalidInSecurityMode signals that the second code is invalid in security mode
var ErrSecondCodeInvalidInSecurityMode = errors.New("second code is invalid in security mode")
