---
title: "Operations"
summary: "Production operations procedures, monitoring, and security."
tags: ["operations", "hub", "index"]
area: "operations"
status: "stable"
owner: "team-ops"
version: "1.0"
last_reviewed: 2025-12-24
aliases: ["ops hub"]
---

# Operations

Production operations, monitoring, security, and maintenance procedures.

## Quick Links

### Kubernetes Operations
- [[kubernetes-runbook|Kubernetes Runbook]] - Complete K8s operations guide
- [[kubernetes-scaling|Kubernetes Scaling]] - HPA, cluster autoscaling, resource management
- [[kubernetes-troubleshooting|Kubernetes Troubleshooting]] - Common issues and debugging
- [[kubernetes-disaster-recovery|Kubernetes Disaster Recovery]] - Backup, restore, failover
- [[backup-recovery-runbook|Backup & Recovery Runbook]] - Automated backups, PITR, restore testing
- [[kubernetes-cost-optimization|Kubernetes Cost Optimization]] - Cost reduction strategies
- [[resource-quotas|Resource Quotas & Limits]] - Quota enforcement and OOM prevention

### Platform Operations
- [[monitoring|Monitoring]] - Metrics, logs, and alerting
- [[preflight|Preflight Checklist]] - Pre-deployment validation
- [[migration|Database Migrations]] - Migration procedures
- [[secrets-management|Secrets Management]] - Secure credential handling
- [[security-scanning|Security Scanning]] - Automated security checks
- [[security-testing-runbook|Security Testing Runbook]] - Security testing procedures
- [[waf-protection|WAF Protection]] - Application-level WAF and rate limiting
- [[ddos-protection|DDoS Protection]] - DDoS mitigation and traffic analytics
- [[observability|Observability]] - Distributed tracing
- [[CDN_FAILOVER_RUNBOOK|CDN Failover Runbook]] - CDN failover procedures
- [[DEPLOYMENT_AUTOMATION|Deployment Automation]] - Automated deployment processes
- [[TWITCH_BAN_UNBAN_TESTING_ROLLOUT_DOCS|Twitch Moderation Rollout]] - Twitch moderation feature rollout

## Deployment & Blue-Green

For deployment procedures, see the **[[../deployment/index|Deployment Hub]]**:

- [[../deployment/docker|Docker Deployment]] - Container-based deployment
- [[../deployment/ci_cd|CI/CD Pipeline]] - GitHub Actions workflows
- [[../deployment/infra|Infrastructure]] - Cloud infrastructure and scaling
- [[../deployment/runbook|Operations Runbook]] - Day-to-day operational procedures
- [[deployment|Deployment Guide]] - Deployment overview
- [[blue-green-deployment|Blue-Green Deployment]] - Zero-downtime deployments
- [[blue-green-rollback|Blue-Green Rollback]] - Rolling back deployments
- [[blue-green-quick-reference|Blue-Green Quick Reference]] - Quick reference guide
- [[twitch-moderation-rollout-plan|Twitch Moderation Rollout Plan]] - Feature rollout planning

## Security & Compliance

- [Security Audit Report](./SECURITY_AUDIT_REPORT.md) - Comprehensive pre-launch security audit
- [Security Audit Executive Summary](./SECURITY_AUDIT_EXECUTIVE_SUMMARY.md) - Executive overview and key findings
- [Security Audit Checklist](./SECURITY_AUDIT_CHECKLIST.md) - Detailed security audit checklist (250 items)
- [[break-glass-procedures|Break Glass Procedures]] - Emergency access procedures
- [[credential-rotation-runbook|Credential Rotation]] - Regular credential rotation

## On-Call Playbooks

- [[playbooks/README|Playbooks Overview]] - Operational playbooks index
- [[playbooks/search-incidents|Search Incidents]] - Semantic search troubleshooting
- [[playbooks/slo-breach-response|SLO Breach Response]] - Responding to SLO violations
- [[runbooks/alert-validation|Alert Validation]] - Validating alert configurations
- [[runbooks/background-jobs|Background Jobs]] - Background job troubleshooting
- [[runbooks/hpa-scaling|HPA Scaling]] - Horizontal pod autoscaling
- [[runbook|Operations Runbook]] - General operational procedures

## Moderation System Runbooks

Comprehensive operational runbooks for managing the moderation system in production:

- [[runbooks/moderation-operations|Moderation Operations]] - Emergency procedures, manual ban/unban, moderator management
- [[runbooks/audit-log-operations|Audit Log Operations]] - Review and export audit logs, compliance procedures
- [[runbooks/ban-sync-troubleshooting|Ban Sync Troubleshooting]] - Twitch ban synchronization issues and solutions
- [[runbooks/permission-escalation|Permission Escalation]] - Grant/revoke permissions, emergency access procedures
- [[runbooks/moderation-rollback|Moderation Rollback]] - Feature flag rollback, database rollback, emergency disable
- [[runbooks/moderation-monitoring|Moderation Monitoring]] - Key metrics, alert configuration, dashboard setup
- [[runbooks/moderation-incidents|Moderation Incidents]] - Common issues, security incidents, contact procedures

## Advanced Topics

- [[feature-flags|Feature Flags]] - Gradual rollouts
- [[feature-flags-guide|Feature Flags Guide]] - Implementation guide
- [[performance|Performance]] - Performance optimization
- [[slos|Service Level Objectives]] - SLO definitions and tracking
- [[staging-environment|Staging Environment]] - Staging deployment and testing
- [[centralized-logging|Centralized Logging]] - Log aggregation and analysis
- [[ci-cd-secrets|CI/CD Secrets]] - Managing deployment secrets
- [[ci-cd-vault-integration|CI/CD Vault Integration]] - Vault integration for CI/CD
- [[cicd|CI/CD Overview]] - CI/CD pipeline overview
- [[quick-start-cicd|CI/CD Quick Start]] - Getting started with CI/CD
- [[webhook-monitoring|Webhook Monitoring]] - Monitoring outbound webhooks
- [[webhook-dlq-replay-runbook|Webhook DLQ Replay]] - Replaying failed webhooks
- [[alert-testing-staging|Alert Testing in Staging]] - Testing alerting rules
- [[on-call-rotation|On-Call Rotation]] - On-call schedule and procedures
- [[on-call-quick-reference|On-Call Quick Reference]] - Quick reference for on-call
- [[deployment-live-development|Live Development Deployment]] - Development deployment
- [[documentation-hosting|Documentation Hosting]] - Hosting documentation site
- [[admin-system-configuration-ui|Admin System Configuration]] - Admin configuration UI
- [[mfa-admin-guide|MFA Admin Guide]] - Multi-factor authentication setup
- [[query-limits|Query Limits]] - API query limits and throttling
- [[global-redundancy-runbook|Global Redundancy]] - Multi-region redundancy

## Documentation Index

```dataview
TABLE title, summary, status, last_reviewed
FROM "docs/operations"
WHERE file.name != "index"
SORT title ASC
```

---

**See also:**
[[../deployment/index|Deployment Hub]] ·
[[../backend/architecture|Backend Architecture]] ·
[[../setup/development|Development Setup]] ·
[[../index|Documentation Home]]
