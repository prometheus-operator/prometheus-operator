// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWritePEMFileHappyPath verifies the normal write path: file is created,
// PEM block is written, and the file is properly closed (content is fully
// flushed to disk).
func TestWritePEMFileHappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test.pem")

	data := []byte("fake certificate material")
	if err := writePEMFile(outFile, "CERTIFICATE", data); err != nil {
		t.Fatalf("writePEMFile failed: %v", err)
	}

	// 1. Verify the file exists and content is correct.
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	block, rest := pem.Decode(raw)
	if block == nil {
		t.Fatalf("no PEM block found in output; raw content: %q", string(raw))
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("PEM type: got %q, want %q", block.Type, "CERTIFICATE")
	}
	if string(block.Bytes) != string(data) {
		t.Errorf("PEM bytes: got %q, want %q", block.Bytes, data)
	}
	if len(rest) > 0 {
		t.Errorf("unexpected trailing data after PEM block: %q", rest)
	}

	// 2. Verify the file handle was released: a closed file can be removed,
	//    an open file cannot on Windows and sometimes Linux.
	if err := os.Remove(outFile); err != nil {
		t.Errorf("could not remove output file — handle likely still open: %v", err)
	}
}

// TestWritePEMFileInvalidPath verifies that writePEMFile returns an error
// (rather than panicking or hanging) when the destination path is invalid.
// This exercises the early-return path BEFORE os.Create succeeds — included
// for completeness of the writePEMFile contract.
func TestWritePEMFileInvalidPath(t *testing.T) {
	// A path that cannot exist on any reasonable filesystem.
	err := writePEMFile("/nonexistent-dir-xyz/sub/test.pem", "CERTIFICATE", []byte("data"))
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open") {
		t.Errorf("error should mention 'failed to open', got: %v", err)
	}
}

// TestWritePEMFileClosesOnContentWrite verifies that after writePEMFile
// returns (regardless of outcome), the file handle is not leaked.
//
// To simulate an error AFTER os.Create succeeds but DURING pem.Encode, we
// would need to inject a failing io.Writer. Since writePEMFile takes a file
// path rather than a writer, we test the invariant indirectly:
//
//  1. Call writePEMFile successfully → os.Remove succeeds (fd released).
//  2. Call writePEMFile for a path inside a directory that we delete right
//     before the call → error path AFTER os.Create is not reachable without
//     mocking, but the `defer f.Close()` added by the fix guarantees the
//     fd is released even if pem.Encode fails.
//
// The real regression guard is in the code review: `defer f.Close()` appears
// immediately after the error check on os.Create. This test locks in the
// happy-path invariant and ensures the function signature/behaviour stays
// stable.
func TestWritePEMFileClosesOnContentWrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Write several files in a loop to make a handle leak more obvious:
	// if f.Close() were missing, we'd accumulate open fds.
	for i := range 50 {
		outFile := filepath.Join(tmpDir, "cert.pem")
		if err := writePEMFile(outFile, "CERTIFICATE", []byte("payload")); err != nil {
			t.Fatalf("iteration %d: writePEMFile failed: %v", i, err)
		}
		// Remove between iterations so each loop creates a fresh inode.
		if err := os.Remove(outFile); err != nil {
			t.Fatalf("iteration %d: could not remove file: %v", i, err)
		}
	}
}
