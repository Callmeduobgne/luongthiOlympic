// Copyright 2024 IBN Network (ICTU Blockchain Network)
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

package ca

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadIdentityCredentials(baseDir string, candidates []string) ([]byte, []byte, error) {
	if baseDir == "" {
		return nil, nil, fmt.Errorf("MSP directory not configured")
	}

	usersDir := filepath.Join(baseDir, "users")
	var identityPath string
	for _, name := range candidates {
		if name == "" {
			continue
		}
		path := filepath.Join(usersDir, name)
		if _, err := os.Stat(path); err == nil {
			identityPath = path
			break
		}
	}

	if identityPath == "" {
		return nil, nil, fmt.Errorf("identity not found in %s", usersDir)
	}

	certPath := filepath.Join(identityPath, "msp", "signcerts", "cert.pem")
	keyDir := filepath.Join(identityPath, "msp", "keystore")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	keyFiles, err := filepath.Glob(filepath.Join(keyDir, "*_sk"))
	if err != nil || len(keyFiles) == 0 {
		return nil, nil, fmt.Errorf("private key not found in %s", keyDir)
	}

	keyPEM, err := os.ReadFile(keyFiles[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	return certPEM, keyPEM, nil
}

func adminCandidatePaths(mspDir, adminUser, mspID string) []string {
	var candidates []string

	if adminUser != "" {
		candidates = append(candidates, adminUser)
		if !strings.Contains(adminUser, "@") && mspID != "" {
			candidates = append(candidates, fmt.Sprintf("%s@%s", normalizeAdminName(adminUser), strings.ToLower(mspID)))
		}
	}

	if mspID != "" {
		candidates = append(candidates, fmt.Sprintf("Admin@%s", mspID))
		candidates = append(candidates, fmt.Sprintf("admin@%s", strings.ToLower(mspID)))
	}

	if org := extractOrgDomain(mspDir); org != "" {
		candidates = append(candidates, fmt.Sprintf("Admin@%s", org))
		candidates = append(candidates, fmt.Sprintf("admin@%s", org))
	}

	return candidates
}

func userCandidatePath(username string) []string {
	if username == "" {
		return []string{}
	}
	return []string{username}
}

func normalizeAdminName(name string) string {
	if name == "" {
		return "admin"
	}
	return strings.TrimSpace(name)
}

func extractOrgDomain(mspDir string) string {
	const token = "peerOrganizations/"
	idx := strings.Index(mspDir, token)
	if idx == -1 {
		return ""
	}
	rest := mspDir[idx+len(token):]
	if rest == "" {
		return ""
	}
	parts := strings.Split(rest, "/")
	return parts[0]
}
