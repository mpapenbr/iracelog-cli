package log

import "testing"

// verifies the helper function the get the best match logger config
func Test_findBestMatch(t *testing.T) {
	type args struct {
		stringsList []string
		query       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "no match", args: args{
			[]string{"a"}, "x",
		}, want: ""},
		{name: "perfect match", args: args{
			[]string{"a"}, "a",
		}, want: "a"},
		{name: "prefix group match", args: args{
			[]string{"a"}, "a.b",
		}, want: "a"},
		{name: "prefix not match", args: args{
			[]string{"a"}, "ab",
		}, want: ""},
		{name: "prefix group match", args: args{
			[]string{"a"}, "a.b",
		}, want: "a"},
		{name: "prefix group match", args: args{
			[]string{"a", "b"}, "a.b",
		}, want: "a"},
		{name: "prefix group match", args: args{
			[]string{"a", "a.b.c"}, "a.b",
		}, want: "a"},
		{name: "regex match", args: args{
			[]string{"a", "a.\\d+"}, "a.1",
		}, want: "a.\\d+"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findBestMatch(tt.args.stringsList, tt.args.query); got != tt.want {
				t.Errorf("findBestMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
