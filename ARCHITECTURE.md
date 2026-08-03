# OpsPilot Architecture

Version: 1.0

Author: OpsPilot Team

---

# Vision

OpsPilot is an AI-powered Engineering Operations Platform.

Our goal is not to build another monitoring dashboard or incident tracker.

Our goal is to reduce the time required to detect, investigate, understand, and resolve production issues.

The platform continuously discovers infrastructure, monitors system health, correlates telemetry, detects incidents, explains root causes, and assists engineers using AI.

---

# Core Principles

Every feature must satisfy at least one of these goals.

✓ Reduce Mean Time To Detect (MTTD)

✓ Reduce Mean Time To Resolve (MTTR)

✓ Reduce Operational Complexity

✓ Improve Infrastructure Visibility

✓ Increase Automation

✓ Enable AI-assisted Operations

If a feature doesn't help achieve one of these goals, it should not be implemented.

---

# Product Philosophy

OpsPilot is NOT:

- Jira
- Trello
- Kubernetes Dashboard
- Monitoring Dashboard
- Ticket Management Software

OpsPilot IS:

An Engineering Operations Platform.

Monitoring, incidents, alerts, deployments, AI, reports, and infrastructure all belong to the same ecosystem.

---

# Domain Model

Organization

└── Teams

└── Projects

└── Resources

└── Monitoring

└── Alerts

└── Incidents

└── AI Assistant

Everything revolves around Projects.

---

# Resource Hierarchy

Project

├── Cluster

│ ├── Namespace

│ │ ├── Deployment

│ │ │ ├── ReplicaSet

│ │ │ │ └── Pods

│ │ ├── Service

│ │ ├── ConfigMap

│ │ ├── Secret

│ │ ├── Job

│ │ └── CronJob

│

├── Database

├── Cache

├── Queue

├── Storage

├── VM

├── Functions

└── External APIs

This hierarchy becomes the source of truth for the platform.

---

# Major Modules

## Projects

Responsible for ownership.

Projects connect infrastructure.

---

## Resource Inventory

Responsible for discovering infrastructure.

Examples:

Pods

Deployments

Namespaces

VMs

Databases

Redis

Kafka

RabbitMQ

Storage

---

## Monitoring

Collects

Metrics

Logs

Events

Health

Telemetry

---

## Alert Engine

Detects problems.

Groups duplicate alerts.

Applies severity.

Triggers incidents.

---

## Incident Engine

Tracks production incidents.

Maintains timeline.

Stores investigation history.

Tracks ownership.

---

## AI Assistant

Understands

Project

Resources

Metrics

Logs

Alerts

Incidents

Runbooks

Documentation

Git History

Provides

Root Cause Analysis

Troubleshooting

Recommendations

Deployment Impact

Natural Language Queries

---

# Monitoring Philosophy

Monitoring must be automatic.

Users should never manually create health dashboards.

OpsPilot should discover infrastructure automatically.

---

# Incident Philosophy

Incidents are outputs.

Users do not create incidents because something "might" be wrong.

Incidents exist because the system detected a real operational problem.

---

# AI Philosophy

AI is not a chatbot.

AI is an Engineering Copilot.

It understands the customer's infrastructure.

Future capabilities:

Explain failures

Detect anomalies

Predict incidents

Suggest fixes

Generate reports

Answer operational questions

---

# Backend Architecture

Handler

↓

Service

↓

Repository

↓

Database

Business logic belongs inside Services.

Repositories never contain business logic.

Handlers never access databases directly.

---

# Frontend Architecture

Pages

↓

Feature Components

↓

Shared Components

↓

Hooks

↓

Services

↓

API

No component should call Axios directly.

React Query owns all server state.

---

# Database Philosophy

Normalize where appropriate.

Use JSON metadata only when schemas vary between providers.

Keep audit fields on every table.

Never duplicate business data.

Prefer soft relationships over excessive joins when scalability matters.

---

# Event Flow

Infrastructure

↓

Monitoring

↓

Alert

↓

Incident

↓

AI Analysis

↓

User Notification

↓

Resolution

---

# Scalability

Future support:

Multiple Organizations

Multiple Teams

Multiple Projects

Multiple Clusters

Multiple Cloud Providers

Hybrid Infrastructure

Self-hosted

SaaS

---

# Cloud Providers

Azure

AWS

Google Cloud

DigitalOcean

On-Prem Kubernetes

---

# Supported Platforms

Kubernetes

Docker

Linux

Windows Servers

Virtual Machines

Containers

Serverless

---

# Security

JWT Authentication

Role Based Access

Project Isolation

Audit Logs

Secrets Encryption

Secure API Access

---

# Future Roadmap

Phase 1

Projects

Resources

Incidents

Alerts

---

Phase 2

Monitoring

Metrics

Logs

Tracing

---

Phase 3

AI Copilot

Root Cause Analysis

Knowledge Base

RAG

---

Phase 4

Automation

Auto-remediation

Runbooks

AI Actions

---

# Coding Standards

Business logic belongs in Services.

React Query for all server state.

Validation through shared schemas.

No duplicated constants.

Reusable UI components.

Accessibility first.

Type safety everywhere.

---

# Definition of Done

A feature is complete only when it includes:

✓ Backend

✓ Frontend

✓ Validation

✓ Error Handling

✓ Loading States

✓ Empty States

✓ Accessibility

✓ Type Safety

✓ Responsive UI

✓ Documentation

✓ Build Passes

✓ Lint Passes

---

# Long-Term Goal

OpsPilot should become an intelligent operations platform capable of understanding an entire production environment and helping engineers operate complex systems with confidence.

Every feature should move the platform toward that vision.