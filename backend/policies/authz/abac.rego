package authz.abac

import future.keywords.if
import future.keywords.in

# Attribute-Based Access Control rules

# High-value batch approval requires admin or high-level inspector
allow if {
    input.request.resource == "batch"
    input.request.action == "approve"
    input.user.roles[_] == "quality:inspector"
    input.user.attributes.certification_level >= 3
    input.resource.attributes.value <= 10000
    input.resource.attributes.status == "pending_approval"
}

# Admin can approve any batch
allow if {
    input.request.resource == "batch"
    input.request.action == "approve"
    input.user.roles[_] in ["org:admin", "system:admin"]
}

# Time-based access control (business hours only)
allow if {
    input.request.resource == "batch"
    input.request.action == "approve"
    input.user.roles[_] == "quality:inspector"
    is_business_hours(input.environment.timestamp)
}

# Location-based access (if needed)
allow if {
    input.request.resource == "transaction"
    input.request.action == "submit"
    input.environment.attributes.country in input.user.attributes.allowed_countries
}

# Maintenance mode - deny all writes
deny if {
    input.request.action in ["create", "update", "delete", "submit"]
    input.environment.attributes.maintenance_mode == true
}

# Helper: Check if current time is within business hours
is_business_hours(timestamp) if {
    hour := time.clock(timestamp)[0]
    hour >= 9
    hour < 17
}

