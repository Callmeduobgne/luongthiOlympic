/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * Validate batch ID format
 */
export function validateBatchId(batchId: string): void {
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
export function validateDate(dateString: string): void {
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
export function validateString(value: string, fieldName: string, maxLength: number = 500): void {
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
export function validatePagination(limit: number, offset: number): void {
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
export function validatePackageId(packageId: string): void {
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
export function validateWeight(weight: number): void {
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
export function validateDateRange(productionDate: string, expiryDate?: string): void {
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



 */
export function validatePackageId(packageId: string): void {
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
export function validateWeight(weight: number): void {
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
export function validateDateRange(productionDate: string, expiryDate?: string): void {
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


