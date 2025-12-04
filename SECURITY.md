# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Security Overview

IBN Network is a blockchain-based traceability system built on Hyperledger Fabric. Security is a critical aspect of the system, especially given the immutable nature of blockchain transactions and the sensitive nature of supply chain data.

### Security Features

- **TLS Encryption:** All network communications use TLS 1.2+
- **JWT Authentication:** Secure token-based authentication for API access
- **API Key Authentication:** Alternative authentication method for service-to-service communication
- **Role-Based Access Control (RBAC):** Fine-grained permissions via OPA policies
- **Certificate Management:** Proper MSP (Membership Service Provider) configuration
- **Input Validation:** Comprehensive validation at API and chaincode levels
- **Audit Logging:** All critical operations are logged for compliance

## Reporting a Vulnerability

We take the security of IBN Network seriously. If you believe you have found a security vulnerability in IBN Network, please report it to us as described below.

**⚠️ Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

Please send an email to **security@ibn.vn** (or your designated security contact email) with the subject line: `[SECURITY] IBN Network Vulnerability Report`

You should receive a response within **24 hours**. If for some reason you do not, please follow up with us to ensure we received your original message.

### What to Include

Your report should include:

- **A detailed description of the vulnerability**
  - What component is affected (backend, frontend, chaincode, network configuration)
  - The potential impact (data breach, unauthorized access, denial of service, etc.)
  
- **Steps to reproduce the vulnerability**
  - Clear, step-by-step instructions
  - Include code snippets or configuration if relevant
  - Specify the environment (Docker version, OS, etc.)
  
- **Potential impact of the vulnerability**
  - What data or systems could be compromised
  - Likelihood of exploitation
  - Potential business impact
  
- **Suggested fix (if any)**
  - If you have ideas for how to fix the issue, please include them
  - This helps us address the issue more quickly

### Our Response

We will:

1. **Acknowledge receipt** of your report within 24 hours
2. **Investigate** all legitimate reports promptly
3. **Keep you updated** on our progress (at least weekly)
4. **Credit you** in our security advisories (unless you prefer to remain anonymous)
5. **Notify you** when the vulnerability is fixed

### Disclosure Policy

- We will work with you to understand and resolve the issue quickly
- We will not disclose the vulnerability publicly until a fix is available
- We will coordinate disclosure with you if you wish to publish your findings
- We will credit you for the discovery (unless you prefer anonymity)

## Security Best Practices

### For Developers

1. **Never commit sensitive data:**
   - API keys, passwords, private keys
   - Real certificates (use test certificates only)
   - Database credentials

2. **Use environment variables:**
   - Store secrets in `.env` files (not committed)
   - Use Docker secrets for production

3. **Validate all inputs:**
   - API endpoints should validate request data
   - Chaincode should validate transaction parameters
   - Frontend should validate user input

4. **Follow principle of least privilege:**
   - Use Admin MSP only when necessary
   - Implement proper RBAC policies
   - Limit network access where possible

5. **Keep dependencies updated:**
   - Regularly update Go modules
   - Update npm packages
   - Monitor for security advisories

### For System Administrators

1. **Certificate Management:**
   - Regenerate certificates if compromised
   - Use strong cryptographic algorithms
   - Rotate certificates regularly
   - Never share private keys

2. **Network Security:**
   - Use TLS for all communications
   - Implement firewall rules
   - Monitor network traffic
   - Keep Fabric network components updated

3. **Access Control:**
   - Use strong passwords for database and services
   - Implement API key rotation
   - Monitor access logs
   - Use OPA policies for fine-grained control

4. **Backup and Recovery:**
   - Regular backups of database
   - Backup crypto material securely
   - Test recovery procedures
   - Document disaster recovery plan

## Known Security Considerations

### Blockchain-Specific

- **Immutable Transactions:** Once committed, transactions cannot be modified
- **Certificate Expiration:** Monitor certificate expiration dates
- **Genesis Block:** Must match current crypto material (TLS certificates)
- **Channel Policies:** Review and update channel policies regularly

### API Security

- **JWT Tokens:** Tokens expire after configured time
- **API Keys:** Format: `ibn_` prefix + 32 hex characters
- **Rate Limiting:** Consider implementing rate limiting for public endpoints
- **CORS:** Configure CORS properly for frontend access

### Infrastructure

- **Docker Security:** Keep Docker and Docker Compose updated
- **Container Images:** Use official images with specific versions
- **Network Isolation:** Use Docker networks to isolate services
- **Volume Security:** Protect sensitive data in Docker volumes

## Security Updates

Security updates will be released as patch versions (e.g., 1.0.1, 1.0.2) and will be documented in:

- [CHANGELOG.md](CHANGELOG.md) - Under "Security" section
- GitHub Security Advisories (if applicable)
- Release notes

## Compliance

IBN Network is designed with compliance in mind:

- **Audit Logging:** All critical operations are logged
- **Data Integrity:** Blockchain ensures immutability
- **Access Control:** RBAC and OPA policies for authorization
- **Encryption:** TLS for data in transit

## Contact

For security-related questions or concerns, please contact:

- **Email:** security@ibn.vn
- **Response Time:** Within 24 hours

---

**Thank you for helping keep IBN Network secure!** 🔒
