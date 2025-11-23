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

package version

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles version management data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new version management repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// VersionTag represents a version tag
type VersionTag struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	TagName           string
	TagType           string
	Description       *string
	IsActive          bool
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// VersionDependency represents a version dependency
type VersionDependency struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	DependencyName    string
	DependencyVersion string
	DependencyType    string
	IsRequired        bool
	Metadata          json.RawMessage
	CreatedAt         time.Time
}

// VersionReleaseNote represents release notes for a version
type VersionReleaseNote struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	Title             string
	Content           string
	ReleaseType       string
	BreakingChanges   []string
	NewFeatures       []string
	BugFixes          []string
	Improvements      []string
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// VersionComparison represents a version comparison
type VersionComparison struct {
	ID                  uuid.UUID
	FromVersionID       uuid.UUID
	ToVersionID         uuid.UUID
	ComparisonType      string
	ChangesSummary      *string
	BreakingChangesCount int
	NewFeaturesCount    int
	BugFixesCount       int
	Metadata            json.RawMessage
	CreatedAt           time.Time
}

// CreateVersionTag creates a new version tag
func (r *Repository) CreateVersionTag(ctx context.Context, tag *VersionTag) error {
	query := `
		INSERT INTO blockchain.version_tags (
			id, chaincode_version_id, tag_name, tag_type,
			description, is_active, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		tag.ID, tag.ChaincodeVersionID, tag.TagName, tag.TagType,
		tag.Description, tag.IsActive, tag.CreatedBy,
	).Scan(&tag.CreatedAt, &tag.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create version tag: %w", err)
	}

	return nil
}

// GetVersionTagsByVersionID retrieves tags for a version
func (r *Repository) GetVersionTagsByVersionID(ctx context.Context, versionID uuid.UUID) ([]*VersionTag, error) {
	query := `
		SELECT id, chaincode_version_id, tag_name, tag_type,
		       description, is_active, created_by, created_at, updated_at
		FROM blockchain.version_tags
		WHERE chaincode_version_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version tags: %w", err)
	}
	defer rows.Close()

	var tags []*VersionTag
	for rows.Next() {
		tag := &VersionTag{}
		err := rows.Scan(
			&tag.ID, &tag.ChaincodeVersionID, &tag.TagName, &tag.TagType,
			&tag.Description, &tag.IsActive, &tag.CreatedBy,
			&tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetVersionByTag retrieves a version by tag name
func (r *Repository) GetVersionByTag(ctx context.Context, chaincodeName, channelName, tagName string) (*uuid.UUID, error) {
	query := `
		SELECT cv.id
		FROM blockchain.chaincode_versions cv
		JOIN blockchain.version_tags vt ON cv.id = vt.chaincode_version_id
		WHERE cv.name = $1 AND cv.channel_name = $2
		  AND vt.tag_name = $3 AND vt.is_active = TRUE
		  AND cv.deleted_at IS NULL
		ORDER BY cv.created_at DESC
		LIMIT 1
	`

	var versionID uuid.UUID
	err := r.db.QueryRow(ctx, query, chaincodeName, channelName, tagName).Scan(&versionID)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("version not found for tag: %s", tagName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version by tag: %w", err)
	}

	return &versionID, nil
}

// CreateVersionDependency creates a new version dependency
func (r *Repository) CreateVersionDependency(ctx context.Context, dep *VersionDependency) error {
	query := `
		INSERT INTO blockchain.version_dependencies (
			id, chaincode_version_id, dependency_name, dependency_version,
			dependency_type, is_required, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at
	`

	var metadataJSON interface{}
	if dep.Metadata != nil {
		metadataJSON = dep.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		dep.ID, dep.ChaincodeVersionID, dep.DependencyName, dep.DependencyVersion,
		dep.DependencyType, dep.IsRequired, metadataJSON,
	).Scan(&dep.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create version dependency: %w", err)
	}

	return nil
}

// GetVersionDependencies retrieves dependencies for a version
func (r *Repository) GetVersionDependencies(ctx context.Context, versionID uuid.UUID) ([]*VersionDependency, error) {
	query := `
		SELECT id, chaincode_version_id, dependency_name, dependency_version,
		       dependency_type, is_required, metadata, created_at
		FROM blockchain.version_dependencies
		WHERE chaincode_version_id = $1
		ORDER BY dependency_type, dependency_name
	`

	rows, err := r.db.Query(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version dependencies: %w", err)
	}
	defer rows.Close()

	var deps []*VersionDependency
	for rows.Next() {
		dep := &VersionDependency{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&dep.ID, &dep.ChaincodeVersionID, &dep.DependencyName, &dep.DependencyVersion,
			&dep.DependencyType, &dep.IsRequired, &metadataJSON, &dep.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version dependency: %w", err)
		}

		if metadataJSON.Valid {
			dep.Metadata = json.RawMessage(metadataJSON.String)
		}

		deps = append(deps, dep)
	}

	return deps, nil
}

// CreateVersionReleaseNote creates a new release note
func (r *Repository) CreateVersionReleaseNote(ctx context.Context, note *VersionReleaseNote) error {
	query := `
		INSERT INTO blockchain.version_release_notes (
			id, chaincode_version_id, title, content, release_type,
			breaking_changes, new_features, bug_fixes, improvements, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		note.ID, note.ChaincodeVersionID, note.Title, note.Content, note.ReleaseType,
		note.BreakingChanges, note.NewFeatures, note.BugFixes, note.Improvements, note.CreatedBy,
	).Scan(&note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create release note: %w", err)
	}

	return nil
}

// GetVersionReleaseNote retrieves release note for a version
func (r *Repository) GetVersionReleaseNote(ctx context.Context, versionID uuid.UUID) (*VersionReleaseNote, error) {
	query := `
		SELECT id, chaincode_version_id, title, content, release_type,
		       breaking_changes, new_features, bug_fixes, improvements,
		       created_by, created_at, updated_at
		FROM blockchain.version_release_notes
		WHERE chaincode_version_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	note := &VersionReleaseNote{}
	err := r.db.QueryRow(ctx, query, versionID).Scan(
		&note.ID, &note.ChaincodeVersionID, &note.Title, &note.Content, &note.ReleaseType,
		&note.BreakingChanges, &note.NewFeatures, &note.BugFixes, &note.Improvements,
		&note.CreatedBy, &note.CreatedAt, &note.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("release note not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get release note: %w", err)
	}

	return note, nil
}

// CompareVersions compares two versions using database function
func (r *Repository) CompareVersions(ctx context.Context, version1, version2 string) (int, error) {
	query := `SELECT blockchain.compare_semantic_versions($1, $2)`

	var result int
	err := r.db.QueryRow(ctx, query, version1, version2).Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to compare versions: %w", err)
	}

	return result, nil
}

// GetLatestVersion gets the latest version for a chaincode
func (r *Repository) GetLatestVersion(ctx context.Context, chaincodeName, channelName string) (*uuid.UUID, error) {
	query := `SELECT blockchain.get_latest_version($1, $2)`

	var versionID uuid.UUID
	err := r.db.QueryRow(ctx, query, chaincodeName, channelName).Scan(&versionID)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no version found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return &versionID, nil
}

// CreateVersionComparison creates a version comparison record
func (r *Repository) CreateVersionComparison(ctx context.Context, comp *VersionComparison) error {
	query := `
		INSERT INTO blockchain.version_comparisons (
			id, from_version_id, to_version_id, comparison_type,
			changes_summary, breaking_changes_count, new_features_count,
			bug_fixes_count, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at
	`

	var metadataJSON interface{}
	if comp.Metadata != nil {
		metadataJSON = comp.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		comp.ID, comp.FromVersionID, comp.ToVersionID, comp.ComparisonType,
		comp.ChangesSummary, comp.BreakingChangesCount, comp.NewFeaturesCount,
		comp.BugFixesCount, metadataJSON,
	).Scan(&comp.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create version comparison: %w", err)
	}

	return nil
}

// GetVersionComparisons retrieves comparisons for a version
func (r *Repository) GetVersionComparisons(ctx context.Context, versionID uuid.UUID) ([]*VersionComparison, error) {
	query := `
		SELECT id, from_version_id, to_version_id, comparison_type,
		       changes_summary, breaking_changes_count, new_features_count,
		       bug_fixes_count, metadata, created_at
		FROM blockchain.version_comparisons
		WHERE from_version_id = $1 OR to_version_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version comparisons: %w", err)
	}
	defer rows.Close()

	var comparisons []*VersionComparison
	for rows.Next() {
		comp := &VersionComparison{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&comp.ID, &comp.FromVersionID, &comp.ToVersionID, &comp.ComparisonType,
			&comp.ChangesSummary, &comp.BreakingChangesCount, &comp.NewFeaturesCount,
			&comp.BugFixesCount, &metadataJSON, &comp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version comparison: %w", err)
		}

		if metadataJSON.Valid {
			comp.Metadata = json.RawMessage(metadataJSON.String)
		}

		comparisons = append(comparisons, comp)
	}

	return comparisons, nil
}

