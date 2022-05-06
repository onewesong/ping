// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// taken from http://golang.org/src/pkg/net/ipraw_test.go

package ping

import "testing"

func TestPinger(t *testing.T) {
	type args struct {
		localIP  string
		remoteIP string
		timeout  int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "case1", args: args{"0.0.0.0", "1.1.1.1", 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Pinger(tt.args.localIP, tt.args.remoteIP, tt.args.timeout); (err != nil) != tt.wantErr {
				t.Errorf("Pinger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
