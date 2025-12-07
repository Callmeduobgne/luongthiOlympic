"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
exports.validateBatchId = validateBatchId;
exports.validateDate = validateDate;
exports.validateString = validateString;
exports.validatePagination = validatePagination;
exports.validatePackageId = validatePackageId;
exports.validateWeight = validateWeight;
exports.validateDateRange = validateDateRange;
/**
 * Validate batch ID format
 */
function validateBatchId(batchId) {
    if (!batchId || batchId.trim().length === 0) {
        throw new Error("Batch ID cannot be empty");
    }
    if (batchId.length > 100) {
        throw new Error("Batch ID must be less than 100 characters");
    }
    // Allow alphanumeric, dash, underscore
    if (!/^[a-zA-Z0-9_-]+$/.test(batchId)) {
        throw new Error("Batch ID can only contain alphanumeric characters, dashes, and underscores");
    }
}
/**
 * Validate date format (YYYY-MM-DD)
 */
function validateDate(dateString) {
    if (!dateString || dateString.trim().length === 0) {
        throw new Error("Date cannot be empty");
    }
    const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
    if (!dateRegex.test(dateString)) {
        throw new Error("Date must be in format YYYY-MM-DD");
    }
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
        throw new Error("Invalid date");
    }
}
/**
 * Validate string input
 */
function validateString(value, fieldName, maxLength = 500) {
    if (!value || value.trim().length === 0) {
        throw new Error(`${fieldName} cannot be empty`);
    }
    if (value.length > maxLength) {
        throw new Error(`${fieldName} must be less than ${maxLength} characters`);
    }
}
/**
 * Validate pagination parameters
 */
function validatePagination(limit, offset) {
    if (limit < 1 || limit > 1000) {
        throw new Error("Limit must be between 1 and 1000");
    }
    if (offset < 0) {
        throw new Error("Offset must be non-negative");
    }
}
/**
 * Validate package ID format
 */
function validatePackageId(packageId) {
    if (!packageId || packageId.trim().length === 0) {
        throw new Error("Package ID cannot be empty");
    }
    if (packageId.length > 100) {
        throw new Error("Package ID must be less than 100 characters");
    }
    // Allow alphanumeric, dash, underscore
    if (!/^[a-zA-Z0-9_-]+$/.test(packageId)) {
        throw new Error("Package ID can only contain alphanumeric characters, dashes, and underscores");
    }
}
/**
 * Validate weight (must be positive number)
 */
function validateWeight(weight) {
    if (weight <= 0) {
        throw new Error("Weight must be greater than 0");
    }
    if (weight > 100000) {
        throw new Error("Weight must be less than 100000 grams (100 kg)");
    }
}
/**
 * Validate date range (expiryDate must be after productionDate)
 */
function validateDateRange(productionDate, expiryDate) {
    if (!expiryDate) {
        return; // Optional field
    }
    const prodDate = new Date(productionDate);
    const expDate = new Date(expiryDate);
    if (isNaN(prodDate.getTime()) || isNaN(expDate.getTime())) {
        throw new Error("Invalid date format");
    }
    if (expDate <= prodDate) {
        throw new Error("Expiry date must be after production date");
    }
}
//# sourceMappingURL=validation.js.map