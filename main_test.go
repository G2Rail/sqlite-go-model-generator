package main

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func Test_formatFieldName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "downcase to upper case",
			args: args{s: "ruby"},
			want: "Ruby",
		},
		{
			name: "under score to captilized camel case",
			args: args{s: "ruby_on_rails"},
			want: "RubyOnRails",
		},
		{
			name: "white space to underscore",
			args: args{s: "Ruby on rails"},
			want: "Ruby_on_rails",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFieldName(tt.args.s); got != tt.want {
				t.Errorf("formatFieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringifyFirstChar(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "replace leading number with the English name",
			args: args{str: "1ruby"},
			want: "one_ruby",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringifyFirstChar(tt.args.str); got != tt.want {
				t.Errorf("stringifyFirstChar() = %v, want %v", got, tt.want)
			}
		})
	}
}
