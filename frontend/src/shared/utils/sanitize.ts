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

