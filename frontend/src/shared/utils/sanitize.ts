/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

/**
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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
 * HTML sanitization utility
 * Note: DOMPurify is not installed yet, so this is a placeholder
 * Install: npm install dompurify @types/dompurify
 */

/**
 * Basic XSS protection - escape HTML characters
 * For production, use DOMPurify instead
 */
export const escapeHtml = (unsafe: string): string => {
  return unsafe
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

/**
 * Sanitize HTML string
 * TODO: Install DOMPurify and implement proper sanitization
 * 
 * Example with DOMPurify:
 * import DOMPurify from 'dompurify'
 * 
 * export const sanitizeHtml = (html: string): string => {
 *   return DOMPurify.sanitize(html, {
 *     ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'p', 'br'],
 *     ALLOWED_ATTR: ['href', 'target'],
 *   })
 * }
 */
export const sanitizeHtml = (html: string): string => {
  // For now, just escape HTML
  // TODO: Implement with DOMPurify
  return escapeHtml(html)
}

