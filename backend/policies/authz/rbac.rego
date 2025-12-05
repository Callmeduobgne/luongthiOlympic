package authz.rbac

import future.keywords.if
import future.keywords.in

# Role-based access control rules

# System Admin - Full access
allow if {
    input.user.roles[_] == "system:admin"
}

# Organization Admin - Org-level access
allow if {
    input.user.roles[_] == "org:admin"
    input.request.scope == "organization"
}

# Supplier - Can create and read batches
allow if {
    input.user.roles[_] == "supplier"
    input.request.resource == "batch"
    input.request.action in ["create", "read", "update"]
    input.request.scope == "organization"
}

# Manufacturer - Can process batches
allow if {
    input.user.roles[_] == "manufacturer"
    input.request.resource == "batch"
    input.request.action in ["read", "update", "process"]
    input.request.scope in ["organization", "channel"]
}

# Distributor - Can ship batches
allow if {
    input.user.roles[_] == "distributor"
    input.request.resource == "batch"
    input.request.action in ["read", "update", "ship"]
    input.request.scope in ["organization", "channel"]
}

# Retailer - Can sell batches
allow if {
    input.user.roles[_] == "retailer"
    input.request.resource == "batch"
    input.request.action in ["read", "sell"]
    input.request.scope in ["organization", "channel"]
}

# Quality Inspector - Can verify and approve
allow if {
    input.user.roles[_] == "quality:inspector"
    input.request.resource == "batch"
    input.request.action in ["read", "verify", "approve", "reject"]
    input.request.scope in ["organization", "channel"]
}

# Consumer - Read-only access
allow if {
    input.user.roles[_] == "consumer"
    input.request.resource == "batch"
    input.request.action == "read"
    input.request.scope == "public"
}

# Analyst - Read and analytics access
allow if {
    input.user.roles[_] == "analyst"
    input.request.resource in ["batch", "transaction", "analytics"]
    input.request.action in ["read", "query", "analyze"]
    input.request.scope in ["organization", "channel"]
}

