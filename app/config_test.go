// +build windows

package app

import "testing"

var unmarshalConfigTests = []struct {
	id       int
	yaml     []byte
	expected *Config
}{
	{1, []byte("%"), nil},
	{2, []byte("---\n"), &Config{}},
	{3, []byte("---\nps_path: C:\\test\\path"), &Config{PSPath: "C:\\test\\path"}},
	{4, []byte("---\nps_opts: -ExecutionPolicy Bypass"), &Config{PSOpts: "-ExecutionPolicy Bypass"}},
	{5, []byte("---\nscripts_path: C:\\test\\path"), &Config{ScriptsPath: "C:\\test\\path"}},
	{6, []byte("---\ncrt_path: C:\\test\\path"), &Config{CrtPath: "C:\\test\\path"}},
	{7, []byte("---\nkey_path: C:\\test\\path"), &Config{KeyPath: "C:\\test\\path"}},
	{8, []byte("---\nps_path: C:/test/path\nps_opts: -InputFormat None\nscripts_path: C:/scripts"), &Config{PSPath: "C:/test/path", PSOpts: "-InputFormat None", ScriptsPath: "C:/scripts"}},
	{9, []byte("---\nps_path: C:/test/path\nps_opts: -InputFormat None\nscripts_path: C:/scripts\nkey_path: C:/ssl/my.key\ncrt_path: C:/ssl/my.crt"), &Config{PSPath: "C:/test/path", PSOpts: "-InputFormat None", ScriptsPath: "C:/scripts", KeyPath: "C:/ssl/my.key", CrtPath: "C:/ssl/my.crt"}},
}

func TestUnmarshalConfig(t *testing.T) {
	for _, tt := range unmarshalConfigTests {
		actual, err := unmarshalConfig(tt.yaml)
		if err != nil && tt.expected != nil {
			t.Errorf("[id: %d] unmarshalConfig: failed with %s", tt.id, err)
			continue
		}
		if err != nil && tt.expected == nil {
			continue
		}
		if actual.PSOpts != tt.expected.PSOpts ||
			actual.PSPath != tt.expected.PSPath ||
			actual.ScriptsPath != tt.expected.ScriptsPath ||
			actual.KeyPath != tt.expected.KeyPath ||
			actual.CrtPath != tt.expected.CrtPath {
			t.Errorf("[id: %d] unmarshalConfig: expected %q, actual %q", tt.id, tt.expected, actual)
		}
	}
}
