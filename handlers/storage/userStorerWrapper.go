package storage

import "github.com/multiversx/multi-factor-auth-go-service/core"

// ArgUserDataStorerWrapper defines the fields needed to create a new storer wrapper
type ArgUserDataStorerWrapper struct {
	Storer     core.ShardedStorageWithIndex
	Marshaller core.Marshaller
}

type userDataStorerWrapper struct {
	marshaller core.Marshaller
	storer     core.ShardedStorageWithIndex
}

func NewUserDataStorerWrapper(args ArgUserDataStorerWrapper) (*userDataStorerWrapper, error) {
	return &userDataStorerWrapper{
		marshaller: args.Marshaller,
		storer:     args.Storer,
	}, nil
}

func (usw *userDataStorerWrapper) Load(key []byte) (*core.OTPInfo, error) {
	otpInfo, err := usw.getFromStorage(key)
	if err != nil {
		return nil, err
	}

	return usw.decrypt(otpInfo)
}

func (usw *userDataStorerWrapper) Save(key []byte, otpInfo *core.OTPInfo) error {
	encryptedOTPInfo, err := usw.encrypt(otpInfo)
	if err != nil {
		return err
	}

	buff, err := usw.marshaller.Marshal(encryptedOTPInfo)
	if err != nil {
		return err
	}

	return usw.storer.Put(key, buff)
}

func (usw *userDataStorerWrapper) getFromStorage(key []byte) (*core.OTPInfo, error) {
	oldOTPInfo, err := usw.storer.Get(key)
	if err != nil {
		return nil, err
	}

	otpInfo := &core.OTPInfo{}
	err = usw.marshaller.Unmarshal(otpInfo, oldOTPInfo)
	if err != nil {
		return nil, err
	}

	return otpInfo, nil
}

// TODO: implement encryption
func (usw *userDataStorerWrapper) decrypt(otpInfo *core.OTPInfo) (*core.OTPInfo, error) {
	return otpInfo, nil
}

func (usw *userDataStorerWrapper) encrypt(otpInfo *core.OTPInfo) (*core.OTPInfo, error) {
	return otpInfo, nil
}

// IsInterfaceNil return true if there is no value under the interface
func (usw *userDataStorerWrapper) IsInterfaceNil() bool {
	return usw == nil
}
