package config

type FileBase struct {
	File string
}

func (f *FileBase) String() string {
	return f.File
}

func (f *FileBase) Set(flagValue string) error {
	if flagValue == "" {
		f.File = ""
		return nil
	}
	f.File = flagValue
	return nil
}
