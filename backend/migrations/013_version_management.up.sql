-- Version Management Schema
-- Enhanced version tracking with semantic versioning and version comparison

-- Version tags table (for semantic versioning)
CREATE TABLE blockchain.version_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    tag_name VARCHAR(100) NOT NULL, -- e.g., 'v1.0.0', 'latest', 'stable'
    tag_type VARCHAR(50) NOT NULL DEFAULT 'version' CHECK (tag_type IN ('version', 'alias', 'custom')),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (chaincode_version_id, tag_name)
);

-- Version dependencies table
CREATE TABLE blockchain.version_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    dependency_name VARCHAR(255) NOT NULL,
    dependency_version VARCHAR(50) NOT NULL,
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'chaincode' CHECK (dependency_type IN ('chaincode', 'library', 'external')),
    is_required BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Version release notes table
CREATE TABLE blockchain.version_release_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    release_type VARCHAR(50) NOT NULL DEFAULT 'patch' CHECK (release_type IN ('major', 'minor', 'patch', 'hotfix')),
    breaking_changes TEXT[],
    new_features TEXT[],
    bug_fixes TEXT[],
    improvements TEXT[],
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Version comparison table (for tracking version relationships)
CREATE TABLE blockchain.version_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    to_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    comparison_type VARCHAR(50) NOT NULL CHECK (comparison_type IN ('upgrade', 'downgrade', 'rollback', 'sidegrade')),
    changes_summary TEXT,
    breaking_changes_count INTEGER DEFAULT 0,
    new_features_count INTEGER DEFAULT 0,
    bug_fixes_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CHECK (from_version_id != to_version_id)
);

-- Indexes for version_tags
CREATE INDEX idx_version_tags_version_id ON blockchain.version_tags(chaincode_version_id);
CREATE INDEX idx_version_tags_tag_name ON blockchain.version_tags(tag_name);
CREATE INDEX idx_version_tags_tag_type ON blockchain.version_tags(tag_type);
CREATE INDEX idx_version_tags_is_active ON blockchain.version_tags(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_version_tags_version_tag ON blockchain.version_tags(chaincode_version_id, tag_name);

-- Indexes for version_dependencies
CREATE INDEX idx_version_dependencies_version_id ON blockchain.version_dependencies(chaincode_version_id);
CREATE INDEX idx_version_dependencies_name ON blockchain.version_dependencies(dependency_name);
CREATE INDEX idx_version_dependencies_type ON blockchain.version_dependencies(dependency_type);
CREATE INDEX idx_version_dependencies_version_name ON blockchain.version_dependencies(chaincode_version_id, dependency_name);

-- Indexes for version_release_notes
CREATE INDEX idx_version_release_notes_version_id ON blockchain.version_release_notes(chaincode_version_id);
CREATE INDEX idx_version_release_notes_release_type ON blockchain.version_release_notes(release_type);
CREATE INDEX idx_version_release_notes_created_at ON blockchain.version_release_notes(created_at DESC);

-- Indexes for version_comparisons
CREATE INDEX idx_version_comparisons_from_version ON blockchain.version_comparisons(from_version_id);
CREATE INDEX idx_version_comparisons_to_version ON blockchain.version_comparisons(to_version_id);
CREATE INDEX idx_version_comparisons_type ON blockchain.version_comparisons(comparison_type);
CREATE INDEX idx_version_comparisons_from_to ON blockchain.version_comparisons(from_version_id, to_version_id);

-- Triggers
CREATE TRIGGER update_version_tags_updated_at BEFORE UPDATE ON blockchain.version_tags
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_version_release_notes_updated_at BEFORE UPDATE ON blockchain.version_release_notes
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to parse semantic version
CREATE OR REPLACE FUNCTION blockchain.parse_semantic_version(version_str VARCHAR)
RETURNS TABLE(major INTEGER, minor INTEGER, patch INTEGER, prerelease TEXT, build TEXT) AS $$
DECLARE
    parts TEXT[];
    version_part TEXT;
    prerelease_part TEXT;
    build_part TEXT;
    version_parts TEXT[];
BEGIN
    -- Split version string by '-' and '+'
    parts := string_to_array(version_str, '-');
    version_part := parts[1];
    
    IF array_length(parts, 1) > 1 THEN
        prerelease_part := parts[2];
        -- Check if there's a build part after '+'
        IF position('+' IN prerelease_part) > 0 THEN
            build_part := substring(prerelease_part FROM position('+' IN prerelease_part) + 1);
            prerelease_part := substring(prerelease_part FROM 1 FOR position('+' IN prerelease_part) - 1);
        END IF;
    END IF;
    
    -- Check for build in version part
    IF position('+' IN version_part) > 0 THEN
        build_part := substring(version_part FROM position('+' IN version_part) + 1);
        version_part := substring(version_part FROM 1 FOR position('+' IN version_part) - 1);
    END IF;
    
    -- Parse version numbers
    version_parts := string_to_array(version_part, '.');
    
    RETURN QUERY SELECT
        COALESCE((version_parts[1]::INTEGER), 0)::INTEGER as major,
        COALESCE((version_parts[2]::INTEGER), 0)::INTEGER as minor,
        COALESCE((version_parts[3]::INTEGER), 0)::INTEGER as patch,
        COALESCE(prerelease_part, '')::TEXT as prerelease,
        COALESCE(build_part, '')::TEXT as build;
END;
$$ LANGUAGE plpgsql;

-- Function to compare semantic versions
CREATE OR REPLACE FUNCTION blockchain.compare_semantic_versions(version1 VARCHAR, version2 VARCHAR)
RETURNS INTEGER AS $$
DECLARE
    v1 RECORD;
    v2 RECORD;
BEGIN
    SELECT * INTO v1 FROM blockchain.parse_semantic_version(version1);
    SELECT * INTO v2 FROM blockchain.parse_semantic_version(version2);
    
    -- Compare major
    IF v1.major > v2.major THEN
        RETURN 1;
    ELSIF v1.major < v2.major THEN
        RETURN -1;
    END IF;
    
    -- Compare minor
    IF v1.minor > v2.minor THEN
        RETURN 1;
    ELSIF v1.minor < v2.minor THEN
        RETURN -1;
    END IF;
    
    -- Compare patch
    IF v1.patch > v2.patch THEN
        RETURN 1;
    ELSIF v1.patch < v2.patch THEN
        RETURN -1;
    END IF;
    
    -- Compare prerelease (empty > non-empty)
    IF v1.prerelease = '' AND v2.prerelease != '' THEN
        RETURN 1;
    ELSIF v1.prerelease != '' AND v2.prerelease = '' THEN
        RETURN -1;
    ELSIF v1.prerelease != '' AND v2.prerelease != '' THEN
        IF v1.prerelease > v2.prerelease THEN
            RETURN 1;
        ELSIF v1.prerelease < v2.prerelease THEN
            RETURN -1;
        END IF;
    END IF;
    
    RETURN 0; -- Equal
END;
$$ LANGUAGE plpgsql;

-- Function to get latest version for a chaincode
CREATE OR REPLACE FUNCTION blockchain.get_latest_version(p_chaincode_name VARCHAR, p_channel_name VARCHAR)
RETURNS UUID AS $$
DECLARE
    v_version_id UUID;
    v_latest_version VARCHAR;
BEGIN
    -- Get the version with highest semantic version
    SELECT id INTO v_version_id
    FROM blockchain.chaincode_versions
    WHERE name = p_chaincode_name
      AND channel_name = p_channel_name
      AND deleted_at IS NULL
    ORDER BY 
        (SELECT major FROM blockchain.parse_semantic_version(version)) DESC,
        (SELECT minor FROM blockchain.parse_semantic_version(version)) DESC,
        (SELECT patch FROM blockchain.parse_semantic_version(version)) DESC,
        committed_at DESC NULLS LAST
    LIMIT 1;
    
    RETURN v_version_id;
END;
$$ LANGUAGE plpgsql;

