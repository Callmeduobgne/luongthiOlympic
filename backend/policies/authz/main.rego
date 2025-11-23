package authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow = false

# Allow if user has required permission
allow if {
    # Get user permissions
    user_permissions := get_user_permissions(input.user.id)
    
    # Check if permission matches
    some permission in user_permissions
    permission_matches(permission, input.request)
    
    # Check conditions (ABAC)
    conditions_satisfied(permission.conditions, input)
}

# Permission matching logic
permission_matches(permission, request) if {
    # Resource match
    resource_matches(permission.resource, request.resource)
    
    # Action match
    action_matches(permission.action, request.action)
    
    # Effect is "allow"
    permission.effect == "allow"
}

# Resource matching
resource_matches(pattern, resource) if {
    # Exact match
    pattern == resource
}

resource_matches(pattern, resource) if {
    # Wildcard match
    pattern == "*"
}

resource_matches(pattern, resource) if {
    # Pattern match (e.g., "batch:*")
    startswith(resource, pattern)
}

# Action matching
action_matches(pattern, action) if {
    pattern == action
}

action_matches(pattern, action) if {
    pattern == "*"
}

# Condition checking
conditions_satisfied(conditions, req_input) if {
    # No conditions = always satisfied
    not conditions
}

conditions_satisfied(conditions, req_input) if {
    # Check user attributes
    user_attributes_match(conditions.user_attributes, req_input.user)
    
    # Check resource attributes
    resource_attributes_match(conditions.resource_attributes, req_input.resource)
    
    # Check environment
    environment_match(conditions.environment_attributes, req_input.environment)
    
    # Check time window
    time_window_match(conditions.time_window, req_input.timestamp)
}

# User attributes matching
user_attributes_match(required, actual) if {
    # Check each required attribute
    every key, value in required {
        compare_attribute(actual[key], value)
    }
}

# Resource attributes matching
resource_attributes_match(required, actual) if {
    # Check each required attribute
    every key, value in required {
        compare_attribute(actual[key], value)
    }
}

# Environment matching
environment_match(required, actual) if {
    # Check each required attribute
    every key, value in required {
        compare_attribute(actual[key], value)
    }
}

# Attribute comparison
compare_attribute(actual, expected) if {
    # Direct equality
    actual == expected
}

compare_attribute(actual, expected) if {
    # Comparison operators
    expected.gte
    actual >= expected.gte
}

compare_attribute(actual, expected) if {
    expected.lte
    actual <= expected.lte
}

compare_attribute(actual, expected) if {
    expected.gt
    actual > expected.gt
}

compare_attribute(actual, expected) if {
    expected.lt
    actual < expected.lt
}

# Time window matching
time_window_match(window, ts) if {
    not window
}

time_window_match(window, ts) if {
    # Parse timestamp (simplified - assumes RFC3339 format)
    # In production, use proper time parsing
    hour := time.clock(ts)[0]
    minute := time.clock(ts)[1]
    current_time := (hour * 60) + minute
    
    start_time := parse_time(window.start_time)
    end_time := parse_time(window.end_time)
    
    current_time >= start_time
    current_time <= end_time
}

# Helper: Parse time string (HH:MM format)
parse_time(time_str) = result if {
    parts := split(time_str, ":")
    hour := to_number(parts[0])
    minute := to_number(parts[1])
    result := (hour * 60) + minute
}

# Helper: Get user permissions (placeholder - will be populated from database)
get_user_permissions(user_id) = permissions if {
    # This will be populated from database via input
    permissions := input.user.permissions
}

get_user_permissions(user_id) = [] if {
    not input.user.permissions
}

